package main

import (
	"context"
	"fmt"

	"go-filecoin-storage-helper/api/client"
	"go-filecoin-storage-helper/impl"
	"go-filecoin-storage-helper/meta"
	"go-filecoin-storage-helper/proxy"
	"go-filecoin-storage-helper/repo"
	"go-filecoin-storage-helper/store"

	"github.com/multiformats/go-multiaddr"
	"golang.org/x/xerrors"
	"gopkg.in/urfave/cli.v2"
)

var proxyCmd = &cli.Command{
	Name:  "proxy",
	Usage: "proxy cmd",
	Subcommands: []*cli.Command{
		runCmd,
		addDealCmd,
		delDealCmd,
		listMinersCmd,
		listAsksCmd,
	},
}
var runCmd = &cli.Command{
	Name:  "run",
	Usage: "Run filecoin storage helper proxy",
	Action: func(cctx *cli.Context) error {
		repopath := cctx.String(StorageHelperRepo)
		r, err := repo.NewFS(repopath)
		if err != nil {
			return err
		}

		if err := r.Init(); err != nil {
			fmt.Println("Initializing storage helper err:", err)
			return err
		}
		cfg, err := r.GetConfig()
		if err != nil {
			return err
		}
		apima, err := multiaddr.NewMultiaddr(cfg.API.ListenAddress)
		if err != nil {
			return err
		}

		nodeApi, err := meta.ApiFactory(cfg.NODEAPI.Type)
		if err != nil {
			return err
		}

		storageDir := r.GetPath(repo.DataStoreDir)

		ctx := context.WithValue(context.Background(), repo.CtxRepoPath, repopath)
		p := &impl.ProxyImpl{
			Px: proxy.NewProxy(
				proxy.NodeApi(nodeApi),
				proxy.StoreDbSource(storageDir),
				proxy.SynerPeriod(cfg.PROXY.SyncerPeriod),
				proxy.Round(cfg.PROXY.ProxyPeriod),
			),
		}

		go p.Px.Run(ctx)

		return serveRPC(p, apima)
	},
}

var addDealCmd = &cli.Command{
	Name:      "add-deal",
	Usage:     "add deal to proxy",
	ArgsUsage: "<filecid>",
	Action: func(cctx *cli.Context) error {
		ctx := context.Background()
		if cctx.Args().Len() != 3 {
			return xerrors.New("'add' expected 1 args: <filecid>")
		}

		fileCid := cctx.Args().Get(0)

		api, closer, err := client.NewProxyRPC("ws://127.0.0.1:6789/rpc/v0", nil)
		if err != nil {
			return err
		}
		defer closer()
		dealMap, err := meta.DealState(ctx, fileCid)
		if err != nil {
			return xerrors.New("not found deal")
		}

		for cid, deal := range dealMap {
			if err := api.AddDealRenew(ctx, deal.DealID, store.TransferState(deal.State), deal.ExpDate, cid); err != nil {
				return xerrors.New("add deal proxy failed")
			}
		}

		return nil
	},
}

var delDealCmd = &cli.Command{
	Name:      "del-deal",
	Usage:     "delete deal from proxy",
	ArgsUsage: "<fileCid>",
	Action: func(cctx *cli.Context) error {
		ctx := context.Background()
		if cctx.Args().Len() != 1 {
			return xerrors.New("'del' expected 1 args: <fileCid>")
		}

		fileCid := cctx.Args().Get(0)
		api, closer, err := client.NewProxyRPC("ws://127.0.0.1:6789/rpc/v0", nil)
		if err != nil {
			return err
		}
		defer closer()

		dealMap, err := meta.DealState(ctx, fileCid)
		if err != nil {
			return xerrors.New("not found deal")
		}

		for _, deal := range dealMap {
			if err := api.DelDealRenew(ctx, deal.DealID); err != nil {
				return xerrors.New("delete deal from proxy failed")
			}
		}

		return nil
	},
}

var listMinersCmd = &cli.Command{
	Name:  "list-miner",
	Usage: "list miner by key ,like power...",
	Flags: []cli.Flag{
		&cli.Uint64Flag{Name: "count", Value: 10},
		&cli.StringFlag{Name: "key", Value: "power"},
	},

	Action: func(cctx *cli.Context) error {
		ctx := context.Background()

		count := cctx.Uint64("count")
		if count < 1 {
			return nil
		}

		key := cctx.String("key")

		api, closer, err := client.NewProxyRPC("ws://127.0.0.1:6789/rpc/v0", nil)
		if err != nil {
			return err
		}
		defer closer()

		ms, err := api.ListMiners(ctx, key, count)
		if err != nil {
			return xerrors.New("get miner list from proxy failed")
		}

		for i, v := range ms {
			fmt.Printf("Sn:%d,\tMiner:%s,\tPower:%s\n", i, v.Addr, v.Power)
		}

		return nil
	},
}

var listAsksCmd = &cli.Command{
	Name:  "list-ask",
	Usage: "list ask by key ,like price...",
	Flags: []cli.Flag{
		&cli.Uint64Flag{Name: "count", Value: 10},
		&cli.StringFlag{Name: "key", Value: "price"},
	},

	Action: func(cctx *cli.Context) error {
		ctx := context.Background()

		count := cctx.Uint64("count")
		if count < 1 {
			return nil
		}

		key := cctx.String("key")

		api, closer, err := client.NewProxyRPC("ws://127.0.0.1:6789/rpc/v0", nil)
		if err != nil {
			return err
		}
		defer closer()

		ms, err := api.ListAsks(ctx, key, count)
		if err != nil {
			return xerrors.New("get ask list from proxy failed")
		}

		for i, v := range ms {
			fmt.Printf("Sn:%d,\tMiner:%s,\taskId:%s,\tprice:%s,\tExpire:%s\n", i, v.Addr, v.AskID, v.Price, v.Expire)
		}

		return nil
	},
}
