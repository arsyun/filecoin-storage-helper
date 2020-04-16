package lotus

import (
	"context"
	"encoding/json"
	"fmt"
	api "go-filecoin-storage-helper/lib/nodeapi"
	"go-filecoin-storage-helper/utils"
	"strings"

	"golang.org/x/xerrors"
)

const (
	rpcVersion = "rpc/v0"
	defaultURL = "http://127.0.0.1:1234"
)

type LotusAPI struct {
	api.BasicAPI
	token string
}

func NewLotusAPI(opts ...api.Option) api.API {
	options := api.NewOptions(opts...)
	l := new(LotusAPI)

	if options.URL == "" {
		l.URL = defaultURL
	} else {
		l.URL = options.URL
	}

	if options.Token == "" {
		l.token, _ = utils.GetToken()
	} else {
		l.token = options.Token
	}

	return l
}

func (l *LotusAPI) Import(ctx context.Context, path string) (string, error) {
	resp := importresult{}
	b := urlParams("ClientImport", "[\""+path+"\"]")

	err := l.Request(rpcVersion).
		Body(strings.NewReader(b)).
		Exec(ctx, &resp)
	if err != nil {
		return "", err
	}

	return resp.Result.Str, nil
}

func (l *LotusAPI) StartDeal(ctx context.Context, cid string, duration int64, miner string, opts ...api.DealOption) (*api.DealInfo, error) {
	options, err := api.DealOptions(opts...)
	if err != nil {
		return nil, err
	}
	price := options.Price

	c := api.Cid{
		Str: cid,
	}
	pcid, _ := json.Marshal(c)
	//(cid, waddr, maddr, price, dur)
	walletaddr, err := l.GetDefaultWallet(ctx)
	if err != nil {
		return nil, err
	}

	params := fmt.Sprintf("[%s,\"%s\",\"%s\",\"%s\",%d]", string(pcid), walletaddr, miner, price, duration)
	b := urlParams("ClientStartDeal", params)

	resp := dealresult{}
	err = l.Request(rpcVersion).
		Body(strings.NewReader(b)).
		Exec(ctx, &resp)

	if err != nil {
		return nil, err
	}

	return &api.DealInfo{
		DealID: resp.Result.Str,
	}, nil
}

func (l *LotusAPI) QueryDeal(ctx context.Context, cid string) (*api.DealStatus, error) {
	dealid := api.Cid{
		Str: cid,
	}
	enc, err := json.Marshal(dealid)
	if err != nil {
		return nil, err
	}

	b := urlParams("ClientGetDealInfo", "["+string(enc)+"]")
	resp := querydealresult{}
	err = l.Request(rpcVersion).
		Body(strings.NewReader(b)).
		Exec(ctx, &resp)

	if err != nil {
		return nil, err
	}
	if resp.Error.Code != 0 {
		return nil, xerrors.New(resp.Error.Message)
	}

	return &api.DealStatus{
		DealID: resp.Result.ProposalCid.Str,
		State:  resp.Result.State.String(),
	}, nil
}

func (l *LotusAPI) RetriveFile(ctx context.Context, miner string, cid string, path string) error {
	offer, err := l.FindData(ctx, cid)
	if err != nil {
		return err
	}

	payer, err := l.GetDefaultWallet(ctx)
	if err != nil {
		return err
	}
	p := api.Address{
		Str: payer,
	}

	par, err := json.Marshal(offer[0].Order(p))
	if err != nil {
		return err
	}

	b := urlParams("ClientRetrive", "[\""+string(par)+","+path+"\"]")
	_, err = l.Request(rpcVersion).
		Body(strings.NewReader(b)).
		Send(ctx)

	if err != nil {
		return err
	}

	return nil

}

//TODO: retrive file, select the best miners from the offers
func (l *LotusAPI) FindData(ctx context.Context, cid string) ([]offer, error) {
	b := urlParams("ClientFindData", "[\""+cid+"\"]")
	resp := findresult{}

	err := l.Request(rpcVersion).
		Body(strings.NewReader(b)).
		Exec(ctx, &resp)

	if err != nil {
		return nil, nil
	}

	return resp.Result, nil
}

func (l *LotusAPI) GetDefaultWallet(ctx context.Context) (string, error) {
	resp := walletdefaultresult{}
	b := urlParams("WalletDefaultAddress", "[]")
	err := l.Request(rpcVersion).
		Body(strings.NewReader(b)).
		Exec(ctx, &resp)

	if err != nil {
		return "", err
	}

	if resp.Error.Code != 0 {
		return "", xerrors.New(resp.Error.Message)
	}

	return resp.Result, nil
}

func (l *LotusAPI) ChainHead(ctx context.Context) (*api.TipSet, error) {
	resp := ChainHeadResult{}
	b := urlParams("ChainHead", "[]")

	err := l.Request(rpcVersion).
		Body(strings.NewReader(b)).
		Exec(ctx, &resp)

	if resp.Error.Code != 0 {
		return nil, xerrors.New(resp.Error.Message)
	}

	if err != nil {
		return nil, err
	}

	return &api.TipSet{
		Height: resp.Result.Height,
	}, nil
}

func (l *LotusAPI) MinerList(ctx context.Context) ([]string, error) {
	resp := MinerListResult{}

	b := urlParams("StateListMiners", "[[]]")
	err := l.Request(rpcVersion).
		Body(strings.NewReader(b)).
		Exec(ctx, &resp)

	if err != nil {
		return []string{}, err
	}

	if resp.Error.Code != 0 {
		return []string{}, xerrors.New(resp.Error.Message)
	}

	return resp.Result, nil
}

func (l *LotusAPI) MinerPower(ctx context.Context, addr string) (string, error) {
	resp := MinerPowerResult{}
	b := urlParams("StateMinerPower", "[\""+addr+"\""+","+"[]]")

	err := l.Request(rpcVersion).
		Body(strings.NewReader(b)).
		Exec(ctx, &resp)
	if err != nil {
		return "", err
	}

	if resp.Error.Code != 0 {
		return "", xerrors.New(resp.Error.Message)
	}

	return resp.Result.MinerPower, nil
}

//Lotus is not yet implemented
func (l *LotusAPI) LiskAsks(ctx context.Context) ([]*api.Ask, error) {
	return nil, nil
}

//TODO: make deal, select the best storage miner from storage market
func (l *LotusAPI) GetBestMiner(ctx context.Context) (string, error) {
	return "", nil
}

func (l *LotusAPI) Request(command string, args ...string) api.RequestBuilder {
	headers := make(map[string]string)
	if l.Headers != nil {
		for k := range l.Headers {
			headers[k] = l.Headers.Get(k)
		}
	}

	headers["Authorization"] = "Bearer " + l.token
	return api.NewRequestBuilder(command, args, headers, l)
}

func urlParams(method, params string) string {
	str := `{"Jsonrpc": "2.0", "method":"Filecoin.` + method + `", "params":` + params + `, "id":1}`
	return str
}
