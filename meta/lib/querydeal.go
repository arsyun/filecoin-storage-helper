package lib

import (
	"context"
	"encoding/json"
	api "go-filecoin-storage-helper/lib/nodeapi"

	dsq "github.com/ipfs/go-datastore/query"
	"github.com/prometheus/common/log"
)

type QueryDeals struct {
	api    api.API
	dealID string
	store  *MetaStore
}

func NewQueryDeal(dealID string, ds *MetaStore, api api.API) *QueryDeals {
	return &QueryDeals{
		api:    api,
		dealID: dealID,
		store:  ds,
	}
}

func (q *QueryDeals) QueryDeal(ctx context.Context) (map[string]*api.DealStatus, error) {
	//key: cid  value: dealid
	res, err := q.store.Query(dsq.Query{})
	if err != nil {
		return nil, err
	}

	statusmap := make(map[string]*api.DealStatus)
	for {
		deals, ok := res.NextSync()
		if !ok {
			break
		}

		dealstatus := &api.DealStatus{}
		if err := json.Unmarshal(deals.Value, &dealstatus); err != nil {
			return nil, err
		}

		if dealstatus.DealID != "" {
			// s, err := q.querydeal(ctx, dealstatus.DealID)
			s, err := q.api.QueryDeal(ctx, dealstatus.DealID)
			if err != nil {
				log.Errorf("querydeal status failed, dealid: %s", dealstatus.DealID)
			}
			dealstatus.State = s.State
		}
		//list all state
		statusmap[string(deals.Key)] = dealstatus
		//update deal status
		denc, _ := json.Marshal(&dealstatus)
		if err := q.store.Put(denc, deals.Key); err != nil {
			log.Errorf("failed update deal status, dealid: %s", deals.Key)
		}
	}

	return statusmap, nil
}
