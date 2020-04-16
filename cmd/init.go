package main

import (
	"fmt"
	"go-filecoin-storage-helper/repo"

	"gopkg.in/urfave/cli.v2"
)

var initCmd = &cli.Command{
	Name:  "init",
	Usage: "Initialize filecoin storage helper repo",
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

		log.Info("Initializing storage helper success")

		return nil
	},
}
