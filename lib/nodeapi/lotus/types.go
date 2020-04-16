package lotus

import (
	"fmt"
	api "go-filecoin-storage-helper/lib/nodeapi"
)

type DealState uint64

const (
	DealUnknown  = DealState(iota)
	DealRejected // Provider didn't like the proposal
	DealAccepted // Proposal accepted, data moved
	DealStaged   // Data put into the sector
	DealSealing  // Data in process of being sealed

	DealFailed
	DealComplete

	// Internal

	DealError // deal failed with an unexpected error

	DealNoUpdate = DealUnknown
)

func (s DealState) String() string {
	switch s {
	case DealUnknown:
		return "unknow"
	case DealRejected:
		return "rejected"
	case DealAccepted:
		return "accepted"
	case DealStaged:
		return "staged"
	case DealSealing:
		return "sealing"
	case DealFailed:
		return "failed"
	case DealComplete:
		return "complete"
	case DealError:
		return "error"
	default:
		return fmt.Sprintf("<unrecognized %d>", s)
	}
}

type basicjsonrpc struct {
	Jsonrpc string    `json:"jsonrpc"`
	ID      uint64    `json:"id"`
	Error   errResult `json:"error"`
}

type errResult struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// Import
type importresult struct {
	basicjsonrpc
	Result api.Cid `json:"result"`
}

//StartDeal
type dealresult struct {
	basicjsonrpc
	Result api.Cid `json:"result"`
}

//QueryDeal

type querydealresult struct {
	basicjsonrpc
	Result dealinfo `json:"result"`
}

type dealinfo struct {
	ProposalCid   api.Cid   `json:"proposalcid"`
	State         DealState `json:"state"`
	Provider      string    `json:"provider"`
	PieceRef      string    `json:"prieceref"`
	Size          uint64    `json:"size"`
	PricePerEpoch string    `json:"priceperepoch"`
	Duration      uint64    `json:"duration"`
}

//Retrivefile

type retrievalOrder struct {
	// TODO: make this less unixfs specific
	Root api.Cid
	Size uint64
	// TODO: support offset
	Total int64

	Client      api.Address
	Miner       api.Address
	MinerPeerID string
}

//FindData
type findresult struct {
	basicjsonrpc
	Result []offer `json:"result"`
}

type offer struct {
	Err         string      `json:"err"`
	Root        api.Cid     `json:"json"`
	Size        uint64      `json:"size"`
	MinPrice    int64       `json:"minprice"`
	Miner       api.Address `json:"miner"`
	MinerPeerID string      `json:"minerpeerid"`
}

func (o *offer) Order(client api.Address) retrievalOrder {
	return retrievalOrder{
		Root:        o.Root,
		Size:        o.Size,
		Total:       o.MinPrice,
		Client:      client,
		Miner:       o.Miner,
		MinerPeerID: o.MinerPeerID,
	}
}

//WalletDefaultAddress
type walletdefaultresult struct {
	basicjsonrpc
	Result string `json:"result"`
}

//chainhead
type ChainHeadResult struct {
	basicjsonrpc
	Result TipSet `json:"result"`
}

type TipSet struct {
	Height uint64 `json:"height"`
}

//minerlist
type MinerListResult struct {
	basicjsonrpc
	Result []string `json:"result"`
}

//minerpower
type MinerPowerResult struct {
	basicjsonrpc
	Result MinerPower `json:"result"`
}

type MinerPower struct {
	MinerPower string `json:"minerpower"`
	TotalPower string `json:"totalpower"`
}
