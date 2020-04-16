package proxy

import (
	"context"
	"fmt"
	"go-filecoin-storage-helper/api"
	"go-filecoin-storage-helper/config"
	"go-filecoin-storage-helper/utils"

	nodeapi "go-filecoin-storage-helper/lib/nodeapi"
	"go-filecoin-storage-helper/store"
	"go-filecoin-storage-helper/syncer"
	"sync"
	"time"

	"golang.org/x/xerrors"
)

const (
	MinerKeyPower  = "power"
	AskKeyPrice    = "price"
	ProxyStorageDb = "storage.db"
)

type DealInfo struct {
	nodeApi nodeapi.API

	deal *api.Deals

	syncer syncer.Syncer
}

func (di *DealInfo) Update(e syncer.Event) error {
	fmt.Println("proxy Update enter,dealId:", di.deal.DealID)
	if e.Height >= di.deal.ExpDate {
		//no api to renew the order, currently can only retrieve the data, and then reorder
		fpath := "/root" + di.deal.Cid
		if err := utils.GenerateFileByPath(fpath); err != nil {
			return err
		}

		if err := di.nodeApi.RetriveFile(context.Background(), "", di.deal.Cid, fpath); err != nil {
			return err
		}

		cid, err := di.nodeApi.Import(context.Background(), fpath)
		if err != nil {
			return err
		}

		st, err := store.NewStorage("", di.syncer, di.nodeApi)

		minerAsk, err := st.MinPriceMiner()
		if err != nil {
			return err
		}

		dealInfo, err := di.nodeApi.StartDeal(context.Background(), cid, int64(config.DefaultDealDuration), minerAsk.Addr, nodeapi.Price(minerAsk.Price), nodeapi.AskID(minerAsk.AskID))
		if err != nil {
			return err
		}

		chainHeight, err := di.nodeApi.ChainHead(context.Background())
		deal := &api.Deals{
			DealID:  dealInfo.DealID,
			Cid:     cid,
			State:   store.TransferState(dealInfo.State),
			ExpDate: chainHeight.Height + uint64(config.DefaultDealDuration),
		}

		if _, err := st.DelDeals(di.deal.DealID); err != nil {
			return err
		}

		if err := st.AddDeals(deal); err != nil {
			return err
		}

		utils.RemoveFileOrDir(fpath)
	}

	return nil
}

var _ syncer.SyncerObserver = &DealInfo{}

type Event struct {
	State     string
	DealID    string
	DealState int
}

type ProxyNotifier interface {
	Attach(ProxyObserver)
	Detach(ProxyObserver)
	Notify(*Event)
}

type ProxyObserver interface {
	Update(Event) error
}

var _ ProxyNotifier = &Proxy{}

type Proxy struct {
	syncer syncer.Syncer

	deals map[string]*DealInfo
	dMut  sync.Mutex

	st *store.Storage

	options Options

	observers map[ProxyObserver]struct{}

	mut sync.Mutex
}

//attach observer
func (p *Proxy) Attach(o ProxyObserver) {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.observers[o] = struct{}{}
	return
}

//detach observer
func (p *Proxy) Detach(o ProxyObserver) {
	p.mut.Lock()
	defer p.mut.Unlock()
	delete(p.observers, o)
	return
}

//notify observers
func (p *Proxy) Notify(e *Event) {
	p.mut.Lock()
	defer p.mut.Unlock()
	for o, _ := range p.observers {
		o.Update(*e)
	}

	return
}

func (p *Proxy) Run(ctx context.Context) error {
	//check deals
	ticker := time.NewTicker(time.Duration(p.options.Round) * config.BlockerDelay)
	for {
		select {
		case <-ticker.C:
			start := time.Now()
			p.dMut.Lock()
			for _, v := range p.deals {
				//get deal's info
				d, err := p.GetDealStatus(ctx, v.deal.DealID)
				if err != nil {
					return err
				}
				//create event
				e := &Event{
					DealID:    v.deal.DealID,
					DealState: d.deal.State,
				}
				//notify observers
				p.Notify(e)
			}
			p.dMut.Unlock()
			dur := time.Since(start)
			fmt.Println("Proxy notify took:", dur)
		}
	}

	return nil
}

func (p *Proxy) GetDealStatus(ctx context.Context, dealId string) (*DealInfo, error) {
	d, err := p.st.GetDealById(dealId)
	if err != nil {
		return nil, err
	}

	return &DealInfo{
		deal: d,
	}, nil
}

//add deal renew observer
func (p *Proxy) AddDealRenew(state int, dealId string, expDate uint64, cid string) error {
	p.st.AddDeals(&api.Deals{
		DealID:  dealId,
		Cid:     cid,
		State:   state,
		ExpDate: expDate,
	})
	return p.addDealRenew(state, dealId, expDate, cid)
}

func (p *Proxy) addDealRenew(state int, dealId string, expDate uint64, cid string) error {
	p.dMut.Lock()
	defer p.dMut.Unlock()
	p.deals[dealId] = &DealInfo{
		nodeApi: p.options.NodeApi,
		syncer:  p.syncer,
		deal: &api.Deals{
			DealID:  dealId,
			Cid:     cid,
			State:   state,
			ExpDate: expDate,
		},
	}

	p.syncer.Attach(p.deals[dealId])

	return nil
}

//delete deal renew observer
func (p *Proxy) DelDealRenew(dealId string) error {
	p.dMut.Lock()
	defer p.dMut.Unlock()

	if _, ok := p.deals[dealId]; ok {
		p.syncer.Detach(p.deals[dealId])
		p.deals[dealId] = nil
		delete(p.deals, dealId)
	}

	p.st.DelDeals(dealId)

	return nil
}

//list miners by key ,like power ...
func (p *Proxy) ListMiners(key string, cnt uint64) ([]*api.Miner, error) {
	if key == MinerKeyPower {
		return p.st.ListMinersbyPower(cnt)
	}
	return nil, xerrors.New("Please enter list miner keywords")
}

//list Asks by key ,like price ...
func (p *Proxy) ListAsks(key string, cnt uint64) ([]*api.Ask, error) {
	if key == AskKeyPrice {
		return p.st.ListAsksbyPrice(cnt)
	}
	return nil, xerrors.New("Please enter list ask keywords")
}

func NewProxy(opts ...Option) *Proxy {
	options := NewOptions(opts...)

	//create syncer
	sc := syncer.New(
		syncer.Period(options.SynerPeriod),
		syncer.NodeApi(options.NodeApi),
	)

	//create store
	stv, err := store.NewStorage(options.StoreDbSource+"/"+ProxyStorageDb, sc, options.NodeApi)
	if err != nil {
		panic(err.Error())
	}

	d := &Proxy{
		observers: make(map[ProxyObserver]struct{}),
		syncer:    sc,
		st:        stv,
		deals:     make(map[string]*DealInfo),
		options:   options,
	}

	if deals, err := d.st.GetDeals(); err == nil {
		for _, v := range deals {
			d.addDealRenew(v.State, v.DealID, v.ExpDate, v.Cid)
		}
	}

	//store init
	d.st.Init()

	//run syncer
	go d.syncer.Run()

	return d
}
