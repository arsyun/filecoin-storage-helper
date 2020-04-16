package lib

import (
	"context"
	"encoding/json"
	"go-filecoin-storage-helper/lib/encypt"
	api "go-filecoin-storage-helper/lib/nodeapi"
	"go-filecoin-storage-helper/repo"
	"io/ioutil"
	"path/filepath"
	"strings"

	dsq "github.com/ipfs/go-datastore/query"
	"github.com/prometheus/common/log"
	"golang.org/x/xerrors"
)

var (
	incommingdeal chan *Deal

	PrivateKey string
)

type Deal struct {
	miner       string
	cid         string
	askID       uint64
	price       string
	duration    int64
	walletAddr  string
	dstore      *MetaStore
	failedDeals map[string]struct{}
	api         api.API
}

func NewDeal(api api.API, miner string, cid string, duration int64, walletaddr string, askId uint64, price string, repopath string) *Deal {
	ds, err := NewFSstore(repopath, repo.DealDir, cid)
	if err != nil {
		return nil
	}

	return &Deal{
		miner:      miner,
		cid:        cid,
		askID:      askId,
		price:      price,
		duration:   duration,
		walletAddr: walletaddr,
		dstore:     ds,
		api:        api,
	}
}

func (d *Deal) StorageDeal(ctx context.Context, cid string) (*api.DealStatus, error) {
	//If have placed an order before,
	//can re-order the failed order
	if d.isReOrder(ctx, cid) {
		// log.Infof("check failed deal, file cid: %s", cid)
		d.handleFailedDeal(ctx)
	} else {
		ds := api.DealStatus{}
		dealenc, _ := json.Marshal(&ds)
		if err := d.dstore.Put(dealenc, cid); err != nil {
			return nil, xerrors.Errorf("failed put cid to store:", err)
		}

		if err := d.getFileCids(ctx, cid); err != nil {
			//delete the dbfile of deal
			// utils.RemoveFileOrDir(dbpath)
			return nil, err
		}

		d.Startdeal(ctx)
	}

	dealresp := &api.DealStatus{}
	if len(d.failedDeals) > 0 {
		dealresp.State = Failed
	} else {
		dealresp.State = Accepted
		// dealresp.DealId = dealid
	}

	return dealresp, nil
}

func (d *Deal) getFileCids(ctx context.Context, cid string) error {
	repopath, ok := ctx.Value(repo.CtxRepoPath).(string)
	if !ok {
		return xerrors.New("ctx value repopath not found")
	}
	destdbpath := filepath.Join(repopath, repo.MetaDir)
	dir, err := ioutil.ReadDir(destdbpath)
	if err != nil {
		return err
	}

	for _, fi := range dir {
		if strings.HasPrefix(fi.Name(), cid) {
			d.flushCidtoDb(ctx, fi.Name())
		}
	}

	return nil
}

func (d *Deal) flushCidtoDb(ctx context.Context, dbname string) error {
	// log.Infof("flushcid to deal : %s", dbname)
	repopath, ok := ctx.Value(repo.CtxRepoPath).(string)
	if !ok {
		return xerrors.New("ctx value repopath not found")
	}

	st, err := NewFSstore(repopath, "meta", dbname)
	if err != nil {
		return err
	}

	encType, err := st.Get(EncTypeKey)
	if err != nil {
		encType = ""
	} else {
		PrivateKey, _ = encypt.GetKeysbyType(encType, repopath)
	}

	files, err := st.Query(dsq.Query{Prefix: ChunkPrefix})
	if err != nil {
		return xerrors.Errorf("qurey file err:", err)
	}

	senc, _ := json.Marshal(api.DealStatus{})
	var key string
	for {
		cid, ok := files.NextSync()
		if !ok {
			break
		}
		key = string(cid.Value)
		// cid := string(cids.Value)
		if encType != "" {
			deckey, _ := encypt.Decyptdata(encType, key, PrivateKey)
			key = string(deckey)
		}
		// fmt.Printf("key: %s,  value: %s\n", cid.Key, string(cid.Value))
		d.dstore.Put(senc, key)
	}

	return nil
}

func (d *Deal) Startdeal(ctx context.Context) error {
	res, err := d.dstore.Query(dsq.Query{})
	if err != nil {
		return err
	}

	for {
		f, ok := res.NextSync()
		if !ok {
			break
		}

		if _, err := d.proposelStorageDeal(ctx, strings.TrimPrefix(f.Key, "/")); err != nil {
			log.Errorf("proposel deal fail err: %w", err)
		}
	}

	return nil
}

var dealfailmap = map[string]int{
	"unset":    1,
	"unknown":  2,
	"rejected": 3,
	"failed":   4,
	"error":    5,
}

func (d *Deal) handleFailedDeal(ctx context.Context) error {
	//check status
	//key: cid  value: dealid
	res, err := d.dstore.Query(dsq.Query{})
	if err != nil {
		return err
	}

	for {
		f, ok := res.NextSync()
		if !ok {
			break
		}

		deal := api.DealStatus{}
		if err := json.Unmarshal(f.Value, &deal); err != nil {
			return err
		}

		if deal.DealID != "" {
			// s, err := d.dstore.querydeal(ctx, d.StorageAPI, deal.DealID)
			s, err := d.api.QueryDeal(ctx, deal.DealID)
			if err != nil {
				log.Errorf("querydeal status failed, dealid: %s", deal.DealID)
			}
			deal.State = s.State
		}

		if _, ok := dealfailmap[deal.State]; ok {
			log.Infof("handler failed deal, dealid: %s", deal.DealID)
			if _, err := d.proposelStorageDeal(ctx, strings.TrimPrefix(f.Key, "/")); err != nil {
				log.Errorf("proposel deal fail: %w", err)
			}
		}
	}

	return nil
}

func (d *Deal) proposelStorageDeal(ctx context.Context, cid string) (string, error) {
	ds := api.DealStatus{}

	resp, err := d.api.StartDeal(ctx, cid, d.duration, d.miner, api.AskID(d.askID), api.Price(d.price), api.WalletAddr(d.walletAddr))
	if err != nil || resp == nil {
		ds.State = Failed
		d.failedDeals[cid] = struct{}{}
	} else {
		ds.State = Accepted
		ds.DealID = resp.DealID
	}

	chainHead, _ := d.api.ChainHead(ctx)
	ds.ExpDate = chainHead.Height + uint64(d.duration)

	dealenc, _ := json.Marshal(&ds)
	if err := d.dstore.Put(dealenc, cid); err != nil {
		return "", xerrors.Errorf("failed put cid to store: %w", err)
	}

	return resp.DealID, nil
}

//whether cid is in the cid store
//if an order has been placed before the description,
//simply re-order the failed order in the store
func (d *Deal) isReOrder(ctx context.Context, cid string) bool {
	flag, err := d.dstore.Has(cid)
	if err != nil {
		return false
	}
	if flag {
		return true
	}

	return false
}
