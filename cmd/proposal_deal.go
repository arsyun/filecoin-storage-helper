package main

import (
	"context"
	"fmt"
	"strconv"

	"go-filecoin-storage-helper/meta"
	"go-filecoin-storage-helper/repo"

	"golang.org/x/xerrors"
	"gopkg.in/urfave/cli.v2"
)

var proposaldealCmd = &cli.Command{
	Name:      "deal",
	Usage:     "make deal file",
	ArgsUsage: "<cid> <miner> <duration>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "vers",
			Usage: "fil or lotus",
			Value: "lotus",
		},
		&cli.Uint64Flag{
			Name:  "askid",
			Usage: "go-filecoin: ID of ask for which to propose a deal",
		},
		&cli.StringFlag{
			Name:  "price",
			Usage: "lotus: price per epoch",
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := context.Background()
		if cctx.Args().Len() != 3 {
			return fmt.Errorf("'deal' expects 3 args, cid, miner and duration")
		}

		cid := cctx.Args().Get(0)
		miner := cctx.Args().Get(1)
		duration, err := strconv.ParseInt(cctx.Args().Get(2), 10, 64)
		if err != nil {
			fmt.Println("duration arg err:")
			return err
		}

		vers := cctx.String("vers")

		repopath := cctx.String(StorageHelperRepo)
		r, err := repo.NewFS(repopath)
		if err != nil {
			return err
		}

		ok, err := r.Exists()
		if err != nil {
			return err
		}
		if !ok {
			return xerrors.New("repo at is not initialized, run 'storagehelper run' to set it up")
		}

		var askid uint64
		var price string
		switch vers {
		case "fil":
			if cctx.IsSet("askid") {
				askid = cctx.Uint64("askid")
			} else {
				return fmt.Errorf("please set the corrent askid")
			}
		case "lotus":
			if cctx.IsSet("price") {
				price = cctx.String("price")
			} else {
				return fmt.Errorf("please set the corrent price")
			}
		default:
			return fmt.Errorf("please set the correct params")
		}

		dealstate, err := meta.MakeDeal(ctx, cid, miner, duration, meta.DealAPI(vers), meta.AskId(askid), meta.Price(price))
		if err != nil {
			return xerrors.Errorf("deal failed: %w", err)
		}

		if dealstate != nil {
			fmt.Println("dealstate:", dealstate.State)
		}

		return nil
	},
}
