package syncer

import (
	nodeapi "go-filecoin-storage-helper/lib/nodeapi"
	lapi "go-filecoin-storage-helper/lib/nodeapi/lotus"
)

//type NewSyncerHandler func(opts ...Option) Syncer

type Options struct {
	//SyncerInst NewSyncerHandler
	Period  uint64
	NodeApi nodeapi.API
}

type Option func(opts *Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		Period:  5,
		NodeApi: lapi.NewLotusAPI(),
	}

	for _, o := range opts {
		o(&options)
	}

	return options
}

func Period(d uint64) Option {
	return func(options *Options) {
		options.Period = d
	}
}

func NodeApi(na nodeapi.API) Option {
	return func(options *Options) {
		options.NodeApi = na
	}
}

//func SyncerInst(f NewSyncerHandler) Option {
//	return func(options *Options) {
//		options.SyncerInst = f
//	}
//
//}
