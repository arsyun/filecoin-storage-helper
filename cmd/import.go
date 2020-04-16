package main

import (
	"context"
	"fmt"
	"go-filecoin-storage-helper/meta"
	"go-filecoin-storage-helper/repo"

	"golang.org/x/xerrors"
	"gopkg.in/urfave/cli.v2"
)

var importCmd = &cli.Command{
	Name:      "import",
	Usage:     "import data",
	ArgsUsage: "<file>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "vers",
			Usage: "fil or lotus",
			Value: "lotus",
		},
		&cli.Uint64Flag{
			Name:  "size",
			Usage: "File size to split",
			Value: 1 << 20,
		},
		&cli.StringFlag{
			Name:  "pwd",
			Usage: "encypt file",
			// Value: "AES",
			Hidden: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := context.Background()
		if cctx.NArg() != 1 {
			return xerrors.New("'import' expected 1 args: filepath")
		}
		absPath := cctx.Args().Get(0)

		slicesize := cctx.Uint64("size")
		// limit file size
		if slicesize < 1<<20 {
			return xerrors.New("split file cannot be less than 1G")
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

		cid, err := meta.Import(ctx, absPath, meta.ImportAPI(vers), meta.SecSize(slicesize))
		if err != nil {
			return xerrors.Errorf("import failed: %w", err)
		}

		fmt.Println("cid:", cid)
		return nil
	},
}

func verifypwd(key []byte) error {
	k := len(key)
	switch k {
	default:
		return xerrors.Errorf("invalid key size %d, set a 16-bit pwd", k)
	case 16, 24, 32:
		break
	}
	return nil
}
