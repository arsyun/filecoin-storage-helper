package lib

import (
	"go-filecoin-storage-helper/repo"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	badger "github.com/ipfs/go-ds-badger"
)

type MetaStore struct {
	ds  datastore.Batching
	mut sync.Mutex
}

func Newstore(destpath string, ns string) (datastore.Batching, error) {
	opts := badger.DefaultOptions
	opts.Truncate = true

	ds, err := badger.NewDatastore(filepath.Join(destpath, ns), &opts)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (m *MetaStore) Put(value interface{}, k ...string) error {
	var val []byte
	switch v := value.(type) {
	case string:
		val = []byte(v)
	case []byte:
		val = v
	case uint64:
		val = []byte(strconv.FormatUint(v, 10))
	case int64:
		val = []byte(strconv.FormatInt(v, 10))
	case int:
		val = []byte(strconv.Itoa(v))
	}

	m.mut.Lock()
	defer m.mut.Unlock()
	key := datastore.KeyWithNamespaces(k)
	if err := m.ds.Put(key, val); err != nil {
		return err
	}

	return nil
}

func (m *MetaStore) Get(k ...string) (string, error) {
	v, err := m.ds.Get(datastore.KeyWithNamespaces(k))
	if err != nil {
		return "", err
	}

	return string(v), nil
}

func (m *MetaStore) Query(q dsq.Query) (dsq.Results, error) {
	return m.ds.Query(q)
}

func (m *MetaStore) Has(key string) (bool, error) {
	return m.ds.Has(datastore.NewKey(key))
}

func (m *MetaStore) Close() {
	m.ds.Close()
}

func NewFSstore(repopath string, dbtype string, ns string) (*MetaStore, error) {
	Fs, err := repo.NewFS(repopath)
	if err != nil {
		return nil, err
	}

	Fs.Path = filepath.Join(Fs.Path, dbtype)
	ds, err := Fs.Datastore(ns)
	if err != nil {
		return nil, err
	}

	return &MetaStore{
		ds: ds,
	}, nil
}

func NewMemstore() (*MetaStore, error) {
	ds, _ := repo.NewMemory().DataStore()

	return &MetaStore{
		ds: ds,
	}, nil
}
