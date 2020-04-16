package store

import (
	"context"
	"database/sql"
	"sync"

	"go-filecoin-storage-helper/api"
	nodeapi "go-filecoin-storage-helper/lib/nodeapi"
	"go-filecoin-storage-helper/syncer"

	_ "github.com/mattn/go-sqlite3"

	"golang.org/x/xerrors"
)

var (
	once      sync.Once
	StoreInst *Storage
)

type Storage struct {
	db      *sql.DB
	syncer  syncer.Syncer
	nodeApi nodeapi.API

	obs      *StoreObserve
	headerLk sync.Mutex
}

type StoreObserve struct {
	api nodeapi.API

	st *Storage
}

var _ syncer.SyncerObserver = &StoreObserve{}

func (s *StoreObserve) Update(e syncer.Event) error {
	for k, v := range e.PowerList {
		m := &api.Miner{
			Addr:  k,
			Power: v,
		}
		if err := s.st.StorageMiners(m); err != nil {
			continue
		}
	}

	//update deal state
	deals, err := s.st.GetDeals()
	if err != nil {
		return nil
	}

	for _, deal := range deals {
		newDeal, err := s.api.QueryDeal(context.Background(), deal.DealID)
		if err != nil {
			continue
		}

		newState := TransferState(newDeal.State)
		if newState != deal.State {
			deal.State = newState
			s.st.AddDeals(deal)
		}
	}

	return nil
}

func NewStorage(dbSource string, syncer syncer.Syncer, nodeApi nodeapi.API) (*Storage, error) {
	once.Do(func() {
		db, err := sql.Open("sqlite3", dbSource)
		if err != nil {
			return
		}

		StoreInst = &Storage{
			db:      db,
			syncer:  syncer,
			nodeApi: nodeApi,
		}

		if err := StoreInst.setup(); err != nil {
			return
		}
	})

	return StoreInst, nil
}

func (st *Storage) setup() error {
	tx, err := st.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
create table if not exists minerinfo
(
		addr text not null
			constraint miner_pk
				primary key,
		power text not null
);

create table if not exists ask
(
		addr text not null,
		askid bigint not null,
		price text not null,
		expire text not null,
		constraint miner_ask_pk
			primary key (addr, askid)
);

create table if not exists deals
(
		dealid text not null
			constraint deal_pk
				primary key,
		cid	text not null,
		state int not null,
		expdate bigint not null
);

`)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (st *Storage) StorageMiners(m *api.Miner) error {
	tx, err := st.db.Begin()
	if err != nil {
		return nil
	}

	stmt, err := tx.Prepare(`replace into minerinfo(addr, power) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(
		m.Addr,
		m.Power,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (st *Storage) StorageAsks(a *api.Ask) error {
	tx, err := st.db.Begin()
	if err != nil {
		return nil
	}

	stmt, err := tx.Prepare(`insert into ask(addr, askid, price, expire) VALUES (?, ?, ?, ?) on conflict do nothing`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(
		a.Addr,
		a.AskID,
		a.Price,
		a.Expire,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (st *Storage) AddDeals(d *api.Deals) error {
	tx, err := st.db.Begin()
	if err != nil {
		return nil
	}

	stmt, err := tx.Prepare(`replace into deals(dealid, cid, state, expdate) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(
		d.DealID,
		d.Cid,
		d.State,
		d.ExpDate,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (st *Storage) DelDeals(dealId string) (int64, error) {
	tx, err := st.db.Begin()
	if err != nil {
		return 0, err
	}

	stmr, err := tx.Prepare(`delete from deals where dealid = ?`)
	if err != nil {
		return 0, err
	}
	defer stmr.Close()

	res, err := stmr.Exec(dealId)

	if err != nil {
		return 0, err
	}

	affect, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affect, tx.Commit()
}

func (st *Storage) GetDeals() ([]*api.Deals, error) {
	rws, err := st.db.Query(`select * from deals`)
	if err != nil {
		return nil, err
	}

	deals := make([]*api.Deals, 0)
	for rws.Next() {
		d := new(api.Deals)
		if err := rws.Scan(&d.DealID, &d.State, &d.ExpDate); err != nil {
			return nil, err
		}

		deals = append(deals, d)
	}
	if rws.Err() != nil {
		return nil, rws.Err()
	}

	return deals, nil
}

func (st *Storage) GetDealById(dealid string) (*api.Deals, error) {
	rws, err := st.db.Query(`select * from deals where dealid = ?`, dealid)
	if err != nil {
		return nil, err
	}

	deals := make([]*api.Deals, 0)
	for rws.Next() {
		d := new(api.Deals)
		if err := rws.Scan(&d.DealID, &d.Cid, &d.State, &d.ExpDate); err != nil {
			return nil, err
		}

		deals = append(deals, d)
	}

	if rws.Err() != nil {
		return nil, rws.Err()
	}

	return deals[0], nil
}

func (st *Storage) Init() {
	st.obs = &StoreObserve{
		api: st.nodeApi,
		st:  st,
	}
	st.syncer.Attach(st.obs)
}

func (st *Storage) MaxPowerMiner() (*api.Miner, error) {
	rws, err := st.db.Query(`select addr, max(cast (power as BIGINT)) from minerinfo`)
	if err != nil {
		return nil, err
	}

	m := make([]*api.Miner, 0)
	for rws.Next() {
		miner := new(api.Miner)
		if err := rws.Scan(&miner.Addr, &miner.Power); err != nil {
			return nil, err
		}

		m = append(m, miner)
	}
	if rws.Err() != nil {
		return nil, rws.Err()
	}

	return m[0], nil
}

func (st *Storage) MinPriceMiner() (*api.Ask, error) {
	rws, err := st.db.Query(`select addr, askid, min(cast (price as BIGINT)), expire from ask`)
	if err != nil {
		return nil, err
	}

	asks := make([]*api.Ask, 0)
	for rws.Next() {
		a := new(api.Ask)
		if err := rws.Scan(&a.Addr, &a.AskID, &a.Price, &a.Expire); err != nil {
			return nil, err
		}

		asks = append(asks, a)
	}

	if rws.Err() != nil {
		return nil, rws.Err()
	}

	return asks[0], nil
}

func (st *Storage) ListMinersbyPower(cnt uint64) ([]*api.Miner, error) {
	if cnt == 0 {
		return nil, xerrors.New("Para cnt cant be zero")
	}
	rws, err := st.db.Query(`select * from minerinfo order by cast(power as BIGINT) desc`)
	if err != nil {
		return nil, err
	}

	m := make([]*api.Miner, 0)
	var i uint64 = 0
	for rws.Next() {
		if i >= cnt {
			break
		}
		miner := new(api.Miner)
		if err := rws.Scan(&miner.Addr, &miner.Power); err != nil {
			return nil, err
		}

		m = append(m, miner)
		i++
	}
	if rws.Err() != nil {
		return nil, rws.Err()
	}

	return m, nil
}

func (st *Storage) ListAsksbyPrice(cnt uint64) ([]*api.Ask, error) {
	if cnt == 0 {
		return nil, xerrors.New("Para cnt cant be zero")
	}
	rws, err := st.db.Query(`select * from ask order by cast (price as BIGINT)`)
	if err != nil {
		return nil, err
	}

	asks := make([]*api.Ask, 0)
	var i uint64 = 0
	for rws.Next() {
		if i >= cnt {
			break
		}
		a := new(api.Ask)
		if err := rws.Scan(&a.Addr, &a.AskID, &a.Price, &a.Expire); err != nil {
			return nil, err
		}

		asks = append(asks, a)
		i++
	}

	if rws.Err() != nil {
		return nil, rws.Err()
	}

	return asks, nil
}

func (st *Storage) Close() error {
	st.syncer.Detach(st.obs)
	return st.db.Close()
}
