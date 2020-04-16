package lotus_test

import (
	"context"
	lapi "go-filecoin-storage-helper/lib/nodeapi/lotus"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImportfile(t *testing.T) {
	//generate file in /root
	filepath := "/root/1.txt"
	f, err := os.Create(filepath)
	defer f.Close()
	require.NoError(t, err)

	_, err = f.Write([]byte("test jsonrpc"))
	require.NoError(t, err)

	a := lapi.NewLotusAPI()
	cid, err := a.Import(context.Background(), filepath)
	require.NoError(t, err)
	t.Log("cid: ", cid)
}

func TestDeal(t *testing.T) {
	// s := `[{"/":"bafkreiebxp4rrlr6fiw5fmrss56pytak63ua5oi4pic6eejbxf32cuhh7e"},"t3wpv7ahg7fcvga7joge4lul4flkonhk5nrlfwbaaf5c6txwd6hblmgymfo2nlo6jyy2m32gtox3uh2kunbs5q", "t0111", "1", 10]`
}

func TestQueryDeal(t *testing.T) {
	dealid := "xxx"

	a := lapi.NewLotusAPI()
	data, err := a.QueryDeal(context.Background(), dealid)
	require.NoError(t, err)

	t.Logf("resp: %+v", data)
}

func TestGetDefaultWallet(t *testing.T) {
	a := lapi.NewLotusAPI()

	walletAddr, err := a.GetDefaultWallet(context.Background())
	require.NoError(t, err)
	t.Log("wallet addr: ", walletAddr)
}

func TestMinerList(t *testing.T) {
	a := lapi.NewLotusAPI()

	miners, err := a.MinerList(context.Background())
	require.NoError(t, err)
	t.Log("miners: ", miners)
}

func TestMinerPower(t *testing.T) {
	a := lapi.NewLotusAPI()

	minerAddr := "xxx"
	minerPower, err := a.MinerPower(context.Background(), minerAddr)
	require.NoError(t, err)
	t.Log("miner power: ", minerPower)
}
