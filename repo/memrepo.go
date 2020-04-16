package repo

import (
	"github.com/ipfs/go-datastore"
)

type MemRepo struct {
	DS datastore.Datastore
}

func NewMemory() *MemRepo {
	return &MemRepo{}
}

func (m *MemRepo) DataStore() (datastore.Batching, error) {
	ms := datastore.NewMapDatastore()
	return ms, nil
}
