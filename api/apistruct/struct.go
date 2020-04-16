package apistruct

import (
	"context"

	"go-filecoin-storage-helper/api"
)

type ProxyStruct struct {
	Internal struct {
		AddDealRenew func(context.Context, string, int, uint64, string) error                `perm:"admin"`
		DelDealRenew func(context.Context, string) error                                     `perm:"admin"`
		ListMiners   func(ctx context.Context, key string, cnt uint64) ([]*api.Miner, error) `perm:"admin"`
		ListAsks     func(ctx context.Context, key string, cnt uint64) ([]*api.Ask, error)   `perm:"admin"`
	}
}

func (c *ProxyStruct) AddDealRenew(ctx context.Context, dealId string, state int, expDate uint64, cid string) error {
	return c.Internal.AddDealRenew(ctx, dealId, state, expDate, cid)
}

func (c *ProxyStruct) DelDealRenew(ctx context.Context, dealId string) error {
	return c.Internal.DelDealRenew(ctx, dealId)
}

func (c *ProxyStruct) ListMiners(ctx context.Context, key string, cnt uint64) ([]*api.Miner, error) {
	return c.Internal.ListMiners(ctx, key, cnt)
}

func (c *ProxyStruct) ListAsks(ctx context.Context, key string, cnt uint64) ([]*api.Ask, error) {
	return c.Internal.ListAsks(ctx, key, cnt)
}

var _ api.Proxy = &ProxyStruct{}
