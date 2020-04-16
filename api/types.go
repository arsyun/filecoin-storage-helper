package api

type Miner struct {
	Addr string
	// Price string
	Power string
}

type Ask struct {
	Addr string
	//报价单id
	AskID  uint64
	Price  string
	Expire string
}

type Deals struct {
	DealID  string
	Cid     string
	State   int
	ExpDate uint64
}
