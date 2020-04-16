package repo

import (
	"github.com/ipfs/go-datastore"
)

type repo interface {
	Datastore(namespace string) (datastore.Batching, error)
}
