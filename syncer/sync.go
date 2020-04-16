package syncer

type Event struct {
	Height    uint64
	PowerList map[string]string
}

type Syncer interface {
	Attach(SyncerObserver)
	Detach(SyncerObserver)
	Notify(*Event)
	Run()
	Stop()
	ChainGetHeight() (uint64, error)
	ChainGetMinerList(uint64) ([]string, error)
	ChainGetMinerPower(string) (string, error)
}

type SyncerObserver interface {
	Update(Event) error
}

func New(opts ...Option) Syncer {
	return NewBasicSyncer(opts...)
}
