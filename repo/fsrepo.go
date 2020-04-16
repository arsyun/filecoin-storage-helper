package repo

import (
	"os"
	"path/filepath"

	"go-filecoin-storage-helper/config"

	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/xerrors"
)

const (
	CtxRepoPath  = "repopath"
	DealDir      = "deal"
	MetaDir      = "meta"
	DbfileDir    = "dbfile"
	FsConfig     = "config.toml"
	DataStoreDir = "datastore"
)

type FsRepo struct {
	Path string
}

func NewFS(path string) (*FsRepo, error) {
	path, err := homedir.Expand(path)
	if err != nil {
		return nil, err
	}

	return &FsRepo{
		Path: path,
	}, nil
}

func (fsr *FsRepo) Exists() (bool, error) {
	_, err := os.Stat(fsr.Path)
	notexist := os.IsNotExist(err)
	if notexist {
		err = nil
	}

	return !notexist, err
}

func (fsr *FsRepo) Init() error {
	exist, err := fsr.Exists()
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	err = os.Mkdir(fsr.Path, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	if err := fsr.initConfig(); err != nil {
		return err
	}

	//generate meta dir
	if err := fsr.generateMetaDir(); err != nil {
		return err
	}

	//generate db dir
	if err := fsr.generateDbDir(); err != nil {
		return err
	}

	//generate deal dir
	if err := fsr.generateDealDir(); err != nil {
		return err
	}

	//generate storage dir
	if err := fsr.generateDataStoreDir(); err != nil {
		return err
	}

	return nil
}

//init config
func (fsr *FsRepo) initConfig() error {
	cfgP := filepath.Join(fsr.Path, FsConfig)
	_, err := os.Stat(cfgP)
	if err == nil {
		// exists
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	c, err := os.Create(cfgP)
	if err != nil {
		return err
	}

	comm, err := config.ConfigComment(config.DefaultStorageMiner())
	if err != nil {
		return xerrors.Errorf("comment: %w", err)
	}
	_, err = c.Write(comm)
	if err != nil {
		return xerrors.Errorf("write config: %w", err)
	}

	if err := c.Close(); err != nil {
		return xerrors.Errorf("close config: %w", err)
	}

	return nil
}

func (fsr *FsRepo) GetConfig() (*config.StorageHelper, error) {
	cfg := &config.StorageHelper{}
	_, err := config.FromFile(filepath.Join(fsr.Path, FsConfig), cfg)
	return cfg, err
}

func (fsr *FsRepo) GetPath(t string) string {
	return filepath.Join(fsr.Path, t)
}

func (fsr *FsRepo) generateMetaDir() error {
	err := os.Mkdir(filepath.Join(fsr.Path, MetaDir), 0755)
	if err != nil && !os.IsExist(err) {
		return xerrors.Errorf("failed generating meta dir: %w", err)
	}

	return nil
}

func (fsr *FsRepo) generateDbDir() error {
	err := os.Mkdir(filepath.Join(fsr.Path, DbfileDir), 0755)
	if err != nil && !os.IsExist(err) {
		return xerrors.Errorf("failed generating dbfile dir: %w", err)
	}

	return nil
}

func (fsr *FsRepo) generateDealDir() error {
	err := os.Mkdir(filepath.Join(fsr.Path, DealDir), 0755)
	if err != nil && !os.IsExist(err) {
		return xerrors.Errorf("failed generating deal dir: %w", err)
	}

	return nil
}

func (fsr *FsRepo) generateDataStoreDir() error {
	err := os.Mkdir(filepath.Join(fsr.Path, DataStoreDir), 0755)
	if err != nil && !os.IsExist(err) {
		return xerrors.Errorf("failed generating storage dir: %w", err)
	}

	return nil
}

func (fsr *FsRepo) Datastore(ns string) (datastore.Batching, error) {
	opts := badger.DefaultOptions
	opts.Truncate = true
	ds, err := badger.NewDatastore(filepath.Join(fsr.Path, ns), &opts)
	if err != nil {
		return nil, err
	}

	return ds, err
}
