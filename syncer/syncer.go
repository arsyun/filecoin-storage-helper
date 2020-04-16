package syncer

import (
	"context"
	"fmt"
	"go-filecoin-storage-helper/config"
	"sync"
	"time"
)

var _ Syncer = &BasicSyncer{}

type BasicSyncer struct {
	options    Options
	lastHeight uint64

	//observers *list.List
	observers map[SyncerObserver]struct{}
	mut       sync.Mutex

	//power cache
	powerCache map[string]string
	pMut       sync.Mutex
}

//attach observer
func (b *BasicSyncer) Attach(o SyncerObserver) {
	b.mut.Lock()
	defer b.mut.Unlock()
	b.observers[o] = struct{}{}
	return
}

//detach observer
func (b *BasicSyncer) Detach(o SyncerObserver) {
	b.mut.Lock()
	defer b.mut.Unlock()
	delete(b.observers, o)
	return
}

//notify observers
func (b *BasicSyncer) Notify(e *Event) {
	b.mut.Lock()
	defer b.mut.Unlock()
	for o, _ := range b.observers {
		o.Update(*e)
	}
	return
}

func (b *BasicSyncer) Run() {
	ticker := time.NewTicker(config.BlockerDelay * time.Duration(b.options.Period))
	for {
		select {
		case <-ticker.C:
			start := time.Now()
			//get current chain height
			h, err := b.ChainGetHeight()
			if err != nil {
				continue
			}
			//when the chain height dsnt change,continue
			if b.lastHeight == h {
				continue
			}
			//update new height
			b.lastHeight = h

			//create event
			e := &Event{
				PowerList: make(map[string]string),
				Height:    h,
			}

			//get miner list from chain
			l, err := b.ChainGetMinerList(h)
			if err != nil {
				continue
			}

			b.pMut.Lock()
			for _, v := range l {
				//get miner's power from chain
				if p, err := b.ChainGetMinerPower(v); err == nil {
					if pc, ok := b.powerCache[v]; ok {
						if p == pc {
							continue
						}
					}
					//update power cache
					b.powerCache[v] = p
					e.PowerList[v] = p
				}
			}
			b.pMut.Unlock()

			//notify observers
			b.Notify(e)
			dur := time.Since(start)
			fmt.Println("Syncer notify took time:", dur, " miner cnt:", len(l))
		}
	}

}

func (b *BasicSyncer) Stop() {
	return
}

//get height from chain
func (b *BasicSyncer) ChainGetHeight() (uint64, error) {
	ts, err := b.options.NodeApi.ChainHead(context.TODO())
	if err != nil {
		return 0, err
	}

	return ts.Height, nil
}

//get miner list from chain
func (b *BasicSyncer) ChainGetMinerList(h uint64) ([]string, error) {
	miners, err := b.options.NodeApi.MinerList(context.TODO())

	if err != nil {
		return []string{}, err
	}

	return miners, nil
}

//get miner's power from chain
func (b *BasicSyncer) ChainGetMinerPower(addr string) (string, error) {
	power, err := b.options.NodeApi.MinerPower(context.TODO(), addr)
	if err != nil {
		return "", err
	}

	return power, nil
}

func NewBasicSyncer(opts ...Option) Syncer {
	options := NewOptions(opts...)

	bw := &BasicSyncer{
		//observers:  list.New(),
		observers:  make(map[SyncerObserver]struct{}),
		powerCache: make(map[string]string),
		options:    options,
	}

	return bw
}
