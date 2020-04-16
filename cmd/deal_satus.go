package main

import (
	"context"
	"fmt"
	"go-filecoin-storage-helper/meta"
	"go-filecoin-storage-helper/repo"

	"golang.org/x/xerrors"
	"gopkg.in/urfave/cli.v2"
)

var dealstatusCmd = &cli.Command{
	Name:      "state",
	Usage:     "query deal status",
	ArgsUsage: "<cid>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "vers",
			Usage: "fil or lotus",
			Value: "lotus",
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := context.Background()
		if cctx.NArg() != 1 {
			return xerrors.New("'import' expected 1 args: cid")
		}

		cid := cctx.Args().Get(0)

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

		statemap, err := meta.DealState(ctx, cid, meta.DealStateAPI(vers))
		if err != nil {
			return err
		}
		for _, v := range statemap {
			fmt.Printf("dealid: %+v\n", v.DealID)
			fmt.Printf("state: %+v\n", v.State)
		}

		return nil
	},
}
