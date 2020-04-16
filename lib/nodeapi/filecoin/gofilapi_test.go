package filecoin_test

import (
	"context"
	api "go-filecoin-storage-helper/lib/nodeapi"
	filapi "go-filecoin-storage-helper/lib/nodeapi/filecoin"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImport(t *testing.T) {
	filepath := "/root/1.txt"
	_, err := os.Create(filepath)
	require.NoError(t, err)

	storageAPI := filapi.NewGoFileAPI()

	cid, err := storageAPI.Import(context.Background(), filepath)
	require.NoError(t, err)
	t.Log("cid: ", cid)
}

func TestStartDeal(t *testing.T) {
	miner := "xxx"
	cid := "xxxxxx"
	time := int64(1024)
	storageAPI := filapi.NewGoFileAPI()

	d, err := storageAPI.StartDeal(context.Background(), cid, time, miner, api.AskID(0))
	require.NoError(t, err)

	t.Log("state: ", d.State)
	t.Log("messgae: ", d.Message)
	t.Log("dealid: ", d.DealID)
}

func TestQueryDeal(t *testing.T) {
	dealId := ""
	storageAPI := filapi.NewGoFileAPI()

	d, err := storageAPI.QueryDeal(context.Background(), dealId)
	require.NoError(t, err)

	t.Log("state: ", d)
}
