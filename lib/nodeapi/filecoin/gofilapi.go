package filecoin

import (
	"bytes"
	"context"
	"encoding/json"
	api "go-filecoin-storage-helper/lib/nodeapi"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
)

const (
	defaultURL = "http://127.0.0.1:3453/api"
)

type GoFilAPI struct {
	api.BasicAPI
}

func NewGoFileAPI(opts ...api.Option) api.API {
	options := api.NewOptions(opts...)
	g := new(GoFilAPI)

	if options.URL == "" {
		g.URL = defaultURL
	} else {
		g.URL = options.URL
	}

	return g
}

func (g *GoFilAPI) Import(ctx context.Context, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	formfile, err := writer.CreateFormFile("file", "")
	if _, err = io.Copy(formfile, f); err != nil {
		return "", err
	}
	defer writer.Close()

	var out api.Cid
	err = g.Request("client/import").
		Header("Content-Type", writer.FormDataContentType()).
		Body(body).
		Exec(ctx, &out)
	if err != nil {
		return "", err
	}

	return out.Str, nil
}

func (g *GoFilAPI) StartDeal(ctx context.Context, cid string, duration int64, miner string, opts ...api.DealOption) (*api.DealInfo, error) {
	options, err := api.DealOptions(opts...)
	if err != nil {
		return nil, err
	}
	askID := strconv.FormatUint(options.AskID, 10)

	var out DealResult
	d := &api.DealInfo{}

	err = g.Request("client/propose-storage-deal").
		Arguments(miner).
		Arguments(cid).
		Arguments(askID).
		Arguments(strconv.FormatInt(duration, 10)).
		Exec(ctx, &out)
	if err != nil {
		return nil, err
	}

	d.DealID = out.ProposalCid.Str
	d.State = out.State.String()
	d.Message = out.Message

	return d, nil
}

func (g *GoFilAPI) QueryDeal(ctx context.Context, dealid string) (*api.DealStatus, error) {
	var out DealResult

	err := g.Request("client/query-storage-deal", dealid).
		Exec(ctx, &out)
	if err != nil {
		return nil, err
	}

	ds := &api.DealStatus{}
	ds.DealID = out.ProposalCid.Str
	ds.State = out.State.String()

	return ds, nil
}

func (g *GoFilAPI) RetriveFile(ctx context.Context, miner string, cid string, path string) error {
	resp, err := g.Request("client/retrive-piece", cid).
		Arguments(miner).
		Arguments(cid).
		Send(ctx)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	defer resp.Close()

	b := new(bytes.Buffer)
	if _, err := io.Copy(b, resp.Output); err != nil {
		return err
	}

	info, err := ioutil.ReadAll(b)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(info); err != nil {
		return err
	}

	return nil
}

func (g *GoFilAPI) GetDefaultWallet(ctx context.Context) (string, error) {
	return "", nil
}

func (g *GoFilAPI) ChainHead(ctx context.Context) (*api.TipSet, error) {
	var out TipSet
	err := g.Request("chain/head").
		Exec(ctx, &out)
	if err != nil {
		return nil, err
	}

	if len(out.Blocks) == 0 {
		return &api.TipSet{
			Height: 0,
		}, nil
	}

	return &api.TipSet{
		Height: out.Blocks[0].Height,
	}, nil
}

func (g *GoFilAPI) MinerList(ctx context.Context) ([]string, error) {
	var Actors []Actor
	resp, err := g.Request("actor/ls").
		Send(ctx)

	if err != nil {
		return []string{}, err
	}
	if resp.Error != nil {
		return []string{}, resp.Error
	}
	defer resp.Close()

	b := new(bytes.Buffer)
	if _, err = io.Copy(b, resp.Output); err != nil {
		return []string{}, err
	}
	a := "[" + strings.TrimRight(strings.ReplaceAll(b.String(), "\n", ","), ",") + "]"
	if err = json.Unmarshal([]byte(a), &Actors); err != nil {
		return []string{}, err
	}

	minerAddrs := make([]string, 0)
	for _, v := range Actors {
		minerAddrs = append(minerAddrs, v.Key.Str)
	}

	return minerAddrs, nil
}

func (g *GoFilAPI) MinerPower(ctx context.Context, addr string) (string, error) {
	var miner MinerStatus
	err := g.Request("miner/status").
		Arguments(addr).
		Exec(ctx, &miner)
	if err != nil {
		return "", err
	}

	return miner.Power, nil
}

func (g *GoFilAPI) LiskAsks(ctx context.Context) ([]*api.Ask, error) {
	var Asks []*api.Ask
	resp, err := g.Request("client/list-asks").
		Send(ctx)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	defer resp.Close()

	b := new(bytes.Buffer)
	if _, err = io.Copy(b, resp.Output); err != nil {
		return nil, err
	}

	s := "[" + strings.TrimRight(strings.ReplaceAll(b.String(), "\n", ","), ",") + "]"
	if err := json.Unmarshal([]byte(s), &Asks); err != nil {
		return nil, err
	}

	return Asks, nil
}

//TODO: select the best storage miner from storage market
func (g *GoFilAPI) GetBestMiner(ctx context.Context) (string, error) {
	return "", nil
}

func NewGoFilAPIService(opts ...api.Option) api.API {
	return NewGoFileAPI(opts...)
}

func (g *GoFilAPI) Request(command string, args ...string) api.RequestBuilder {
	headers := make(map[string]string)
	if g.Headers != nil {
		for k := range g.Headers {
			headers[k] = g.Headers.Get(k)
		}
	}

	return api.NewRequestBuilder(command, args, headers, g)
}
