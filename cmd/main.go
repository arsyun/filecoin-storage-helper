package main

import (
	"os"

	logging "github.com/ipfs/go-log"

	"gopkg.in/urfave/cli.v2"
)

var log = logging.Logger("main")

const StorageHelperRepo = "storagehelperrepo"

func main() {
	logging.SetLogLevel("*", "INFO")

	local := []*cli.Command{
		initCmd,
		importCmd,
		proposaldealCmd,
		dealstatusCmd,
		retriveCmd,
		proxyCmd,
	}
	app := &cli.App{
		Name:    "filecoin-storage-helper",
		Usage:   " storage file/directory, make deal",
		Version: "v0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  StorageHelperRepo,
				Value: "/root/.storagehelper",
			},
		},
		Commands: local,
	}

	app.Setup()
	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		return
	}
}
