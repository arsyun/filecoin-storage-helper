package meta_test

import (
	"context"
	"testing"

	"go-filecoin-storage-helper/meta"

	"github.com/stretchr/testify/require"
)

func TestMetaImport(t *testing.T) {
	ctx := context.Background()
	err := meta.InitRepo(ctx, "/root/.storagehelper")
	require.NoError(t, err)

	files := "/root/xxx.txt"
	cid, err := meta.Import(ctx, files)
	require.NoError(t, err)
	t.log("cid: ", cid)
}

func TestMetaDeal(t *testing.T) {
	ctx := context.Background()
	err := meta.InitRepo(ctx, "/root/.storagehelper")
	require.NoError(t, err)

	files := "/root/xxx.txt"
	cid, err := meta.Import(context.Background(), files)
	require.NoError(t, err)

	miner := "xxxx"
	dealstate, err := meta.MakeDeal(context.Background(), cid, miner, 120)
	require.NoError(t, err)
	t.log("dealid:", dealstate.DealID)
	t.Log("state：", dealstate.State)
}
