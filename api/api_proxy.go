package api

import (
	"context"
)

type Proxy interface {
	AddDealRenew(context.Context, string, int, uint64, string) error
	DelDealRenew(context.Context, string) error
	ListMiners(ctx context.Context, key string, cnt uint64) ([]*Miner, error)
	ListAsks(ctx context.Context, key string, cnt uint64) ([]*Ask, error)
}
