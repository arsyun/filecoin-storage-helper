package main

import (
	"context"
	"os"

	"go-filecoin-storage-helper/meta"
	"go-filecoin-storage-helper/repo"
	"go-filecoin-storage-helper/utils"

	"golang.org/x/xerrors"
	"gopkg.in/urfave/cli.v2"
)

var retriveCmd = &cli.Command{
	Name:      "retrive",
	Usage:     "retrive file",
	ArgsUsage: "<cid> <targetpath>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "vers",
			Usage: "fil or lotus",
			Value: "LOTUS",
		},
		&cli.StringFlag{
			Name:  "miner",
			Usage: "retrieval miner actor address",
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := context.Background()
		if cctx.NArg() != 2 {
			return xerrors.New("expected 2 args: cid, targetpath")
		}

		cid := cctx.Args().Get(0)

		targetpath := cctx.Args().Get(1)
		if !utils.Exists(targetpath) {
			os.MkdirAll(targetpath, 0755)
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

		var miner string
		if vers == "FIL" {
			if cctx.IsSet("miner") {
				//TODO: automatic retrieval of the most suitable miner
				miner = cctx.String("miner")
			} else {
				return xerrors.New("FIL: must specify retrive miner")
			}
		}

		if err := meta.Retrive(ctx, cid, targetpath, meta.RetrieveAPI(vers), meta.RetrieveMiner(miner)); err != nil {
			return xerrors.Errorf("retrivefile err: %w", err)
		}

		return nil
	},
}
