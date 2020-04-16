package filecoin

import (
	"fmt"
	api "go-filecoin-storage-helper/lib/nodeapi"
)

type State int

const (
	Unset = State(iota)
	Unknown
	Rejected
	Accepted
	Started
	Failed
	Staged
	Complete
)

func (s State) String() string {
	switch s {
	case Unset:
		return "unset"
	case Unknown:
		return "unknown"
	case Rejected:
		return "rejected"
	case Accepted:
		return "accepted"
	case Started:
		return "started"
	case Failed:
		return "failed"
	case Staged:
		return "staged"
	case Complete:
		return "complete"
	default:
		return fmt.Sprintf("<unrecognized %d>", s)
	}
}

type DealResult struct {
	State       State   `json:"state"`
	Message     string  `json:"message"`
	ProposalCid api.Cid `json:"proposalcid"`
}

//chainhead
type TipSet struct {
	Blocks []*Block `json:"blocks"`
}

type Block struct {
	Height uint64 `json:"height"`
}

//minerpower
type MinerStatus struct {
	Power string `json:"power"`
}

//minerlist
type Actor struct {
	Key   api.Address `json:"key"`
	Error error       `json:"error"`
}

//listask
type Ask struct {
	Miner  api.Address
	Price  string
	Expiry uint64
	ID     uint64

	Error error
}
