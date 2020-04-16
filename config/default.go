package config

type StorageHelper struct {
	API     API
	NODEAPI NODEAPI
	PROXY   PROXY
}

type API struct {
	ListenAddress string
}

type NODEAPI struct {
	Type string
}

type PROXY struct {
	SyncerPeriod uint64
	ProxyPeriod  uint64
}

func DefaultStorageMiner() *StorageHelper {
	cfg := &StorageHelper{
		API: API{
			ListenAddress: "/ip4/127.0.0.1/tcp/6789",
		},
		NODEAPI: NODEAPI{
			Type: "lotus",
		},
		PROXY: PROXY{
			SyncerPeriod: 6,
			ProxyPeriod:  10,
		},
	}

	return cfg
}
