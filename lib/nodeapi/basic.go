package api

import (
	"context"
	"net/http"
	gohttp "net/http"
)

type BasicAPI struct {
	URL  string
	Httpcli gohttp.Client
	Headers http.Header
}

func (a *BasicAPI) Import(ctx context.Context, path string) (string, error) {
	return "", nil
}

func (a *BasicAPI) StartDeal(ctx context.Context, cid string, duration int64, miner string, opts ...DealOption) (*DealInfo, error) {
	return nil, nil
}

func (a *BasicAPI) QueryDeal(ctx context.Context, path string) (*DealStatus, error) {
	return nil, nil
}

func (a *BasicAPI) RetriveFile(ctx context.Context, miner string, cid string, path string) error {
	return nil
}

func (a *BasicAPI) GetDefaultWallet(ctx context.Context) (string, error) {
	return "", nil
}

func (a *BasicAPI) ChainHead(ctx context.Context) (*TipSet, error) {
	return nil, nil
}

func (a *BasicAPI) MinerList(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (a *BasicAPI) MinerPower(ctx context.Context, addr string) (string, error) {
	return "", nil
}

func (a *BasicAPI) LiskAsks(ctx context.Context) ([]*Ask, error) {
	return nil, nil
}

func (a *BasicAPI) GetBestMiner(ctx context.Context) (string, error) {
	return "", nil
}

func (a *BasicAPI) GetURL(ctx context.Context) string {
	return a.URL
}

func (a *BasicAPI) GetClient(ctx context.Context) *gohttp.Client {
	return &a.Httpcli
}

var _ API = &BasicAPI{}
