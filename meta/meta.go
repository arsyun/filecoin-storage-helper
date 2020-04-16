package meta

import (
	"context"
	"go-filecoin-storage-helper/lib/encypt"
	api "go-filecoin-storage-helper/lib/nodeapi"
	metalib "go-filecoin-storage-helper/meta/lib"
	"go-filecoin-storage-helper/repo"
	"go-filecoin-storage-helper/utils"
	"path/filepath"

	"golang.org/x/xerrors"
)

//default repopath
var Repopath = "/root/.storagehelper"

//init repo
func InitRepo(ctx context.Context, repopath string) error {
	r, err := repo.NewFS(repopath)
	if err != nil {
		return err
	}

	if err := r.Init(); err != nil {
		return xerrors.Errorf("Init  storagehelper repo err: %w", err)
	}

	return nil
}

func Import(ctx context.Context, file string, opts ...ImportOption) (string, error) {
	if !utils.Exists(Repopath) {
		return "", xerrors.New("must init repo")
	}

	options, err := ImportOptions(opts...)
	if err != nil {
		return "", err
	}

	sliceSize := options.size

	vers := options.vers
	api, err := ApiFactory(vers)
	if err != nil {
		return "", err
	}

	m, err := metalib.NewMetaData(file, repo.MetaDir, api)
	if err != nil {
		return "", err
	}

	pwd := options.pwd
	if pwd != "" {
		//generate keys, default aes
		if err := encypt.GenerateKeyInfo(Repopath, "AES", pwd); err != nil {
			return "", xerrors.New("generate aes key failed")
		}
		m.EncType = "AES"
	}
	m.SliceSize = sliceSize
	//todo: a better way to name it
	m.DbName = utils.GenerateDBName(m.AbsPath) + "_0"

	//Determine whether the metadb already exists, if any, delete
	utils.RemoveFileOrDir(filepath.Join(Repopath, repo.MetaDir, m.DbName))

	s, err := metalib.NewFSstore(Repopath, repo.MetaDir, m.DbName)
	if err != nil {
		return "", err
	}
	m.Mstore = s

	ctx = context.WithValue(ctx, repo.CtxRepoPath, Repopath)
	cid, err := m.Import(ctx)
	if err != nil {
		return "", err
	}

	return cid, nil
}

func MakeDeal(ctx context.Context, cid string, miner string, duration int64, opts ...DealOption) (*api.DealStatus, error) {
	if !utils.Exists(Repopath) {
		return nil, xerrors.New("must init repo")
	}

	options, err := DealOptions(opts...)
	if err != nil {
		return nil, err
	}
	//check param
	vers := options.vers
	if vers == "fil" {
		if options.askId == 0 {
			return nil, xerrors.New("fil: must set askid opts")
		}
	} else if options.price == "" {
		return nil, xerrors.New("lotus: must set price opts")
	}

	api, err := ApiFactory(vers)
	if err != nil {
		return nil, err
	}

	d := metalib.NewDeal(api, miner, cid, duration, "", options.askId, options.price, Repopath)

	ctx = context.WithValue(ctx, repo.CtxRepoPath, Repopath)
	dealState, err := d.StorageDeal(ctx, cid)
	if err != nil {
		return nil, xerrors.New("deal failed")
	}

	if dealState == nil {
		return nil, xerrors.Errorf("file not found: ", err)
	}

	return dealState, nil
}

func DealState(ctx context.Context, cid string, opts ...DealStateOption) (map[string]*api.DealStatus, error) {
	if !utils.Exists(Repopath) {
		return nil, xerrors.New("must init repo")
	}

	options, err := DealStateOptions(opts...)
	if err != nil {
		return nil, err
	}

	api, err := ApiFactory(options.vers)
	if err != nil {
		return nil, err
	}

	dealStore, err := metalib.NewFSstore(Repopath, "deal", cid)
	if err != nil {
		return nil, err
	}

	q := metalib.NewQueryDeal(cid, dealStore, api)
	ctx = context.WithValue(ctx, repo.CtxRepoPath, Repopath)
	statemap, err := q.QueryDeal(ctx)
	// statemap, err := q.QueryDealStatus(ctx, api)
	if err != nil {
		return nil, err
	}

	return statemap, nil
}

func Retrive(ctx context.Context, cid string, targetPath string, opts ...RetrieveOption) error {
	if !utils.Exists(Repopath) {
		return xerrors.New("must init repo")
	}

	options, err := RetrieveOptions(opts...)
	if err != nil {
		return err
	}

	vers := options.vers
	if vers == "fil" {
		if options.miner == "" {
			return xerrors.New("must set retrieve miner")
		}
	}

	api, err := ApiFactory(vers)
	if err != nil {
		return err
	}

	retrive := metalib.NewRetrive(api, options.miner)
	ctx = context.WithValue(ctx, repo.CtxRepoPath, Repopath)
	if err := retrive.RetriveFile(ctx, options.miner, cid, targetPath); err != nil {
		return xerrors.Errorf("retrivefile err: %w", err)
	}

	return nil
}
