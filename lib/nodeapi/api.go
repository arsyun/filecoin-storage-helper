package api

import (
	"context"
	gohttp "net/http"
)

type API interface {
	Import(ctx context.Context, path string) (string, error)

	StartDeal(ctx context.Context, cid string, duration int64, miner string, opts ...DealOption) (*DealInfo, error)

	QueryDeal(ctx context.Context, dealid string) (*DealStatus, error)

	RetriveFile(ctx context.Context, miner string, cid string, path string) error

	GetDefaultWallet(ctx context.Context) (string, error)

	ChainHead(ctx context.Context) (*TipSet, error)

	MinerList(ctx context.Context) ([]string, error)

	MinerPower(ctx context.Context, addr string) (string, error)

	LiskAsks(ctx context.Context) ([]*Ask, error)

	//TODO: select the best miners to place orders
	GetBestMiner(ctx context.Context) (string, error)

	GetURL(ctx context.Context) string

	GetClient(ctx context.Context) *gohttp.Client
}

type Address struct {
	Str string `json:"/"`
}

type Cid struct {
	Str string `json:"/"`
}

type DealInfo struct {
	State   string `json:"state"`
	Message string `json:"message"`
	DealID  string `json:"dealid"`
}

type TipSet struct {
	Height uint64 `json:"height"`
}

type TipSetKey struct {
	Value string `json:"value"`
}

type Ask struct {
	//报价单id
	Askid  uint64
	Maddr  string
	Price  string
	Expire string
}

type DealStatus struct {
	DealID  string `json:"dealid"`
	State   string `json:"state"`
	ExpDate uint64 `json:"expdate"`
}

type APIService interface {
	NewAPI() API
}

func NewAPI(opts ...Option) API {
	options := NewOptions(opts...)

	return options.APIInstance(opts...)
}
