package lib

import (
	"context"
	"go-filecoin-storage-helper/lib/encypt"
	api "go-filecoin-storage-helper/lib/nodeapi"
	pkg "go-filecoin-storage-helper/lib/packing"
	"go-filecoin-storage-helper/repo"
	"go-filecoin-storage-helper/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	dsq "github.com/ipfs/go-datastore/query"
	"github.com/prometheus/common/log"
	"golang.org/x/xerrors"
)

func NewMetaData(fpath string, t string, api api.API) (*MetaData, error) {
	m := &MetaData{}
	m.AbsPath = fpath
	m.Type = t
	m.api = api
	m.EncType = ""
	m.dbNumb = 1
	return m.getFileInfo()
}

type MetaData struct {
	//file/dir name
	Name string

	DbName string
	Size   uint64

	//type: meta; db
	Type string
	//enctype: aes, rsa...
	EncType string
	AbsPath string
	IsDir   bool

	//total num of chunks;
	slices int
	wg     sync.WaitGroup
	Mstore *MetaStore

	//file size
	SliceSize uint64
	Miner     string
	api       api.API
	// Vers       string
	dbNumb int
}

func (m *MetaData) getFileInfo() (*MetaData, error) {
	abspath, err := filepath.Abs(m.AbsPath)
	if err != nil {
		return nil, err
	}

	fr, err := os.Stat(abspath)
	if err != nil {
		if os.IsNotExist(err) {
			return m, err
		}
	}

	m.AbsPath = abspath
	m.Name = fr.Name()
	if fr.IsDir() {
		m.IsDir = true
		dirsize, err := m.dirSize()
		if err != nil {
			return m, err
		}
		m.Size = uint64(dirsize)
	} else {
		m.IsDir = false
		m.Size = uint64(fr.Size())

	}

	return m, nil
}

type chunk struct {
	sequence     int
	absPath      string
	tempfilepath string
}

func (m *MetaData) dirSize() (int64, error) {
	var size int64
	err := filepath.Walk(m.AbsPath, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})

	return size, err
}

func (m *MetaData) Import(ctx context.Context) (string, error) {
	repopath, ok := ctx.Value(repo.CtxRepoPath).(string)
	if !ok {
		return "", xerrors.New("ctx value repopath not found")
	}
	//get aes key
	if m.EncType != "" {
		AESKEY, _ = encypt.GetKeysbyType(m.EncType, repopath)
	}
	//generate temp file dir
	if err := os.MkdirAll(TempFiledir, 0755); err != nil {
		return "", xerrors.Errorf("generate tempfile dir failed: %+v", err)
	}

	m.Type = MetaType
	if err := m.traverseFile(ctx); err != nil {
		return "", err
	}

	cid, err := m.importFile(ctx)
	if err != nil {
		return "", err
	}

	if err := renamedbfile(repopath, m.DbName, cid); err != nil {
		return "", err
	}

	return cid, nil
}

func (m *MetaData) traverseFile(ctx context.Context) error {
	if !m.IsDir {
		if err := m.Mstore.Put(m.Size, FilePrefix, m.AbsPath); err != nil {
			return xerrors.Errorf("saving path to metastore: %w", err)
		}
	} else {
		err := filepath.Walk(m.AbsPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			// key: /path/{abspath}; value: size
			if err := m.Mstore.Put(info.Size(), FilePrefix, path); err != nil {
				return xerrors.Errorf("saving path to metastore: %w", err)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	if err := m.computeSlice(); err != nil {
		return xerrors.Errorf("compute slice err: %w", err)
	}
	if err := m.Mstore.Put(m.slices, SlicesKey); err != nil {
		return xerrors.Errorf("saving slices to metastore: %w", err)
	}
	if err := m.Mstore.Put(m.Type, TypeKey); err != nil {
		return xerrors.Errorf("saving type to metastore: %w", err)
	}
	if err := m.Mstore.Put(m.EncType, EncTypeKey); err != nil {
		return xerrors.Errorf("saving enctype to metastore: %w", err)
	}
	abspath := strings.TrimSuffix(m.AbsPath, m.Name)
	if err := m.Mstore.Put(abspath, AbsPathKey); err != nil {
		return xerrors.Errorf("saving abspath to metastore: %w", err)
	}
	if err := m.Mstore.Put(m.SliceSize, ChunkSizeKey); err != nil {
		return xerrors.Errorf("saving chunksize to metastore: %w", err)
	}

	return nil
}

func (m *MetaData) importFile(ctx context.Context) (string, error) {
	files, err := m.Mstore.Query(dsq.Query{Prefix: FilePrefix})
	if err != nil {
		return "", xerrors.Errorf("query file from metastore err: %w", err)
	}

	m.wg.Add(m.slices)
	for {
		f, ok := files.NextSync()
		if !ok {
			break
		}
		//f.Key: "/path/{abspath}"
		fpath := strings.TrimPrefix(f.Key, FilePrefix)
		fsize, err := strconv.ParseUint(string(f.Value), 10, 64)
		if err != nil {
			return "", xerrors.New("invalid filesize")
		}
		//lotus stores the full file
		if fsize > m.SliceSize {
			go m.sliceFile(ctx, fpath, fsize, &m.wg)
			continue
		}

		chunk := &chunk{
			sequence:     1,
			absPath:      fpath,
			tempfilepath: fpath,
		}

		go m.importChunk(ctx, chunk, &m.wg)
	}
	m.wg.Wait()
	m.Mstore.Close()

	cid, err := m.handlerDb(ctx)
	if err != nil {
		return "", err
	}

	return cid, nil
}

func (m *MetaData) handlerDb(ctx context.Context) (string, error) {
	dbgzpath, err := m.compressDb(ctx)
	if err != nil {
		return "", xerrors.Errorf("compressDb err: %w", err)
	}
	m.Mstore.Close()

	dbgzinfo, _ := os.Stat(dbgzpath)
	dbgzsize := dbgzinfo.Size()

	var cid string
	if uint64(dbgzsize) > m.SliceSize {
		dbmeta, err := m.generateDbMeta(ctx, dbgzpath)
		if err != nil {
			return "", xerrors.Errorf("generateDbMeta err:", err)
		}

		if err := dbmeta.traverseFile(ctx); err != nil {
			return "", err
		}

		return dbmeta.importFile(ctx)
	}

	cid, err = m.importDb(ctx)
	if err != nil {
		return "", xerrors.Errorf("importDb failed: %w", err)
	}

	return cid, nil
}

func (m *MetaData) generateDbMeta(ctx context.Context, fpath string) (*MetaData, error) {
	repopath, ok := ctx.Value(repo.CtxRepoPath).(string)
	if !ok {
		return nil, xerrors.New("ctx value repopath not found")
	}

	dbmeta, err := NewMetaData(fpath, DbType, m.api)
	if err != nil {
		return nil, xerrors.Errorf("failed to new meta instance: %w", err)
	}

	dbName := m.generateDBName()
	s, err := NewFSstore(repopath, MetaType, dbName)
	if err != nil {
		return nil, xerrors.Errorf("failed to new metastore instance: %w", err)
	}

	dbmeta.EncType = m.EncType
	dbmeta.Mstore = s
	dbmeta.DbName = dbName
	dbmeta.SliceSize = m.SliceSize
	dbmeta.AbsPath = fpath
	dbmeta.dbNumb = m.dbNumb + 1

	return dbmeta, err
}

//compress metastore
func (m *MetaData) compressDb(ctx context.Context) (string, error) {
	repopath, ok := ctx.Value(repo.CtxRepoPath).(string)
	if !ok {
		return "", xerrors.New("ctx value repopath not found")
	}

	destFilePath := filepath.Join(repopath, repo.DbfileDir, m.DbName+".tar.gz")
	if utils.Exists(destFilePath) {
		if err := utils.RemoveFileOrDir(destFilePath); err != nil {
			return "", xerrors.New("compressDb remove zip err")
		}
	}

	srcDirPath := filepath.Join(repopath, repo.MetaDir, m.DbName)
	if err := pkg.TarGz(srcDirPath, destFilePath); err != nil {
		return "", xerrors.Errorf("Targz failed: %w", err)
	}

	return destFilePath, nil
}

func (m *MetaData) sliceFile(ctx context.Context, fpath string, fsize uint64, wg *sync.WaitGroup) error {
	fi, err := os.OpenFile(fpath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}

	num := utils.ComputeChunks(fsize, m.SliceSize)
	b := make([]byte, m.SliceSize)
	var i int64 = 1
	for ; i <= int64(num); i++ {

		fi.Seek((i-1)*(int64(m.SliceSize)), 0)

		if len(b) > int((int64(fsize) - (i-1)*int64(m.SliceSize))) {
			b = make([]byte, int64(fsize)-(i-1)*int64(m.SliceSize))
		}

		fi.Read(b)
		//TODO: a better way to name it
		//the split data is generated into temporary files
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		fname := timestamp + strings.Replace(fpath, "/", "_", -1) + strconv.FormatInt(i, 10)

		//generate tempdir
		tempfilepath := filepath.Join(TempFiledir, m.Name, fname)
		if err := utils.GenerateFileByPath(tempfilepath); err != nil {
			return err
		}

		f, err := os.OpenFile(tempfilepath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			log.Errorf("generate tempfile failed, fpath: %s, err: %+v", fpath, err)
		}

		f.Write(b)
		f.Close()
		c := &chunk{
			sequence:     int(i),
			absPath:      fpath,
			tempfilepath: tempfilepath,
		}
		go m.importChunk(ctx, c, wg)
	}

	return nil
}

func (m *MetaData) importChunk(ctx context.Context, c *chunk, wg *sync.WaitGroup) (string, error) {
	defer wg.Done()

	cid, err := m.api.Import(ctx, c.tempfilepath)

	if err != nil {
		return "", xerrors.Errorf("failed to call import api: %w", err)
	}

	//TODO: import data success, delete tempfile
	//Because the source file is retrieved according to cid when placing an order,
	//the source file cannot be deleted here, otherwise the order will fail
	//so can't delete tempfile now

	// if c.absPath != c.tempfilepath {
	// 	if err := os.Remove(c.tempfilepath); err != nil {
	// 		log.Warnf("delete tempfile failed, fpath: %s", c.tempfilepath)
	// 	}
	// }

	cPath := m.chunkPath(c.absPath)
	//checks if set encryption
	if m.EncType != "" {
		cid, err = encypt.Encyptdata(m.EncType, cid, AESKEY)
	}

	if err := m.Mstore.Put(cid, ChunkPrefix, cPath, strconv.Itoa(c.sequence)); err != nil {
		return "", xerrors.Errorf("saving chunkinfo to store: %w", err)
	}

	return cid, nil
}

//if abspath=“C:/a”, name:“a”, c.path:"C:/a/1.txt" then cPath: a/1.txt
func (m *MetaData) chunkPath(c string) string {
	absPrefix := strings.TrimSuffix(m.AbsPath, m.Name)
	cPath := strings.TrimPrefix(c, absPrefix)

	return cPath
}

func (m *MetaData) importDb(ctx context.Context) (string, error) {
	repopath, ok := ctx.Value(repo.CtxRepoPath).(string)
	if !ok {
		return "", xerrors.New("ctx value repopath not found")
	}

	fileName := filepath.Join(repopath, repo.DbfileDir, m.DbName+".tar.gz")
	cid, err := m.api.Import(ctx, fileName)
	if err != nil {
		return "", err
	}

	return cid, nil
}

//Calculate the total number of slices of files/directories
func (m *MetaData) computeSlice() error {
	//query {key: /path/{abspath}}
	files, err := m.Mstore.Query(dsq.Query{Prefix: FilePrefix})
	if err != nil {
		return xerrors.Errorf("qurey file err:", err)
	}

	for {
		f, ok := files.NextSync()
		if !ok {
			break
		}

		fsize, err := strconv.ParseUint(string(f.Value), 10, 64)
		if err != nil {
			return xerrors.New("invalid filesize")
		}

		num := 1
		if fsize > m.SliceSize {
			num = utils.ComputeChunks(fsize, m.SliceSize)
		}

		m.addSlices(num)
	}

	return nil
}

func (m *MetaData) generateDBName() string {
	//m.dbname, eg: root_a_1 or root_a_2 ...
	s := m.DbName[:len(m.DbName)-1]
	return s + strconv.Itoa(m.dbNumb)
}

func (m *MetaData) addSlices(num int) {
	m.slices = m.slices + num
	return
}

func renamedbfile(repopath string, oldName string, newName string) error {
	dbpath := filepath.Join(repopath, repo.MetaDir)
	dir, err := ioutil.ReadDir(dbpath)
	if err != nil {
		return err
	}
	//oldname: root_1.txt_1; root_b.json_2
	//prefixname: root_1.txt
	//newname: cid_1; cid_2
	prefixName := oldName[:len(oldName)-2]

	for _, fi := range dir {
		if strings.HasPrefix(fi.Name(), prefixName) {
			old := filepath.Join(dbpath, fi.Name())
			new := filepath.Join(dbpath, strings.Replace(fi.Name(), prefixName, newName, -1))
			if err = os.Rename(old, new); err != nil {
				return err
			}
		}
	}

	return nil
}
