package impl

import (
	"context"
	"go-filecoin-storage-helper/api"
	"go-filecoin-storage-helper/proxy"
)

type ProxyImpl struct {
	Px *proxy.Proxy
}

func (p *ProxyImpl) AddDealRenew(ctx context.Context, dealId string, state int, expDate uint64, cid string) error {
	return p.Px.AddDealRenew(state, dealId, expDate, cid)
}

func (p *ProxyImpl) DelDealRenew(ctx context.Context, dealId string) error {
	return p.Px.DelDealRenew(dealId)
}

func (p *ProxyImpl) ListMiners(ctx context.Context, key string, cnt uint64) ([]*api.Miner, error) {
	return p.Px.ListMiners(key, cnt)
}

func (p *ProxyImpl) ListAsks(ctx context.Context, key string, cnt uint64) ([]*api.Ask, error) {
	return p.Px.ListAsks(key, cnt)
}

var _ api.Proxy = &ProxyImpl{}
