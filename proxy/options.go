package proxy

import (
	api "go-filecoin-storage-helper/lib/nodeapi"
	//lotussyncer "go-filecoin-storage-helper/syncer/plugins/lotus"
)

//proxy config
type Options struct {
	NodeApi api.API
	//Monitoring period
	Round uint64

	RetryCount int

	RepeatOrder bool

	//SyncerInstHandler syncer.NewSyncerHandler

	SynerPeriod uint64

	Api string

	StoreDbSource string
}

type Option func(*Options)

func NewOptions(opts ...Option) Options {
	options := DefaultOptions()
	for _, o := range opts {
		o(&options)
	}

	return options
}

func DefaultOptions() Options {
	return Options{
		Round:       3,
		RetryCount:  3,
		RepeatOrder: false,
		//SyncerInstHandler: lotussyncer.NewLotusSyner,
		SynerPeriod:   5,
		Api:           "6789",
		StoreDbSource: "/root/.storagehelper/storage/storage.db",
	}
}

func Api(a string) Option {
	return func(options *Options) {
		options.Api = a
	}
}

func NodeApi(api api.API) Option {
	return func(options *Options) {
		options.NodeApi = api
	}
}

func Round(d uint64) Option {
	return func(options *Options) {
		options.Round = d
	}
}

func RetryCount(c int) Option {
	return func(options *Options) {
		options.RetryCount = c
	}
}

func RepeatOrder(r bool) Option {
	return func(options *Options) {
		options.RepeatOrder = r
	}
}

func SynerPeriod(d uint64) Option {
	return func(options *Options) {
		options.SynerPeriod = d
	}
}

func StoreDbSource(dbs string) Option {
	return func(options *Options) {
		options.StoreDbSource = dbs
	}
}
