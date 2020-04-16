package lib

import (
	"context"
	"go-filecoin-storage-helper/lib/encypt"
	api "go-filecoin-storage-helper/lib/nodeapi"
	pkg "go-filecoin-storage-helper/lib/packing"
	"go-filecoin-storage-helper/repo"
	"go-filecoin-storage-helper/utils"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"golang.org/x/xerrors"
)

type Retrive struct {
	//retrive miner
	Miner      string
	StorageAPI api.API
	Mstore     *MetaStore
}

func NewRetrive(api api.API, miner string) *Retrive {
	return &Retrive{
		StorageAPI: api,
		Miner:      miner,
	}
}

func (r *Retrive) RetriveFile(ctx context.Context, miner string, cid string, destpath string) error {
	dbgzpath := filepath.Join(destpath, cid+".tar.gz")
	err := r.StorageAPI.RetriveFile(ctx, miner, cid, dbgzpath)
	if err != nil {
		return err
	}

	if err := r.reductionDb(ctx, dbgzpath, destpath); err != nil {
		return err
	}

	return nil
}

func (r *Retrive) reductionDb(ctx context.Context, dbgzpath string, destpath string) error {
	dirName, err := pkg.UnTarGz(dbgzpath, destpath)
	if err != nil {
		return xerrors.Errorf("UnTarGz failed: %w", err)
	}

	if err := utils.RemoveFileOrDir(dbgzpath); err != nil {
		return err
	}

	ds, err := Newstore(destpath, dirName)
	if err != nil {
		return err
	}

	metaType, err := ds.Get(datastore.NewKey(TypeKey))
	if err != nil {
		return err
	}

	m := &MetaStore{
		ds: ds,
	}

	var destfilepath string
	if string(metaType) == DbType {
		destfilepath = dbgzpath
	} else {
		destfilepath = destpath
	}

	if err := m.generateFile(ctx, r.StorageAPI, r.Miner, destfilepath); err != nil {
		return err
	}
	m.Close()

	if err := utils.RemoveFileOrDir(filepath.Join(destpath, dirName)); err != nil {
		return err
	}

	if string(metaType) == DbType {
		r.reductionDb(ctx, dbgzpath, destpath)
	}

	return nil
}

func (ds *MetaStore) generateFile(ctx context.Context, api api.API, miner string, destpath string) error {
	repopath, ok := ctx.Value(repo.CtxRepoPath).(string)
	if !ok {
		return xerrors.New("ctx value repopath not found")
	}

	cs, err := ds.Get(ChunkSizeKey)
	if err != nil {
		return xerrors.Errorf("get chunksize: %w", err)
	}
	chunksize, _ := strconv.ParseUint(cs, 10, 64)

	abspath, err := ds.Get(AbsPathKey)
	if err != nil {
		return err
	}

	metaType, err := ds.Get(TypeKey)
	if err != nil {
		return err
	}
	//gets the file encryption
	encType, err := ds.Get(EncTypeKey)
	if err != nil {
		encType = ""
	} else {
		PrivateKey, _ = encypt.GetKeysbyType(encType, repopath)
	}

	paths, err := ds.Query(dsq.Query{Prefix: FilePrefix})
	if err != nil {
		return xerrors.Errorf("query fileprefix: %w", err)
	}

	for {
		p, ok := paths.NextSync()
		if !ok {
			break
		}
		// p.Key:"/path/C:/a/b/1.txt"; abspath:"/C:/a/";  fpath: b/1.txt
		fpath := strings.TrimPrefix(p.Key, FilePrefix+string(abspath))
		fsize, err := strconv.ParseUint(string(p.Value), 10, 64)
		if err != nil {
			return xerrors.New("invalid filesize")
		}

		chunknum := utils.ComputeChunks(fsize, chunksize)
		var targetpath string
		if metaType == MetaType {
			targetpath = filepath.Join(destpath, fpath)
			if err := utils.GenerateFileByPath(targetpath); err != nil {
				return xerrors.Errorf("failed generating source file path", err)
			}
		} else {
			targetpath = destpath
		}

		ds.generatechildfile(ctx, api, miner, targetpath, fpath, chunknum, encType)
	}

	return nil
}

func (ds *MetaStore) generatechildfile(ctx context.Context, api api.API, miner string, targetpath string, chunkname string, num int, e string) error {
	for i := 1; i <= num; i++ {
		//Restore the key in the badger database;
		cid, err := ds.Get(ChunkPrefix, chunkname, strconv.Itoa(i))
		if err != nil {
			return err
		}

		if e != "" {
			decid, _ := encypt.Decyptdata(e, cid, PrivateKey)
			cid = string(decid)
		}
		if err := api.RetriveFile(ctx, miner, cid, targetpath); err != nil {
			return err
		}
	}

	return nil
}
