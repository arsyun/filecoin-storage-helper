package store_test

import (
	"go-filecoin-storage-helper/api"
	"go-filecoin-storage-helper/store"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaxPowerMiner(t *testing.T) {
	st, err := store.NewStorage("./storagehelper.db")
	require.NoError(t, err)

	m1 := &api.Miner{
		Addr:  "t0001",
		Power: "122000",
	}

	m2 := &api.Miner{
		Addr:  "t0002",
		Power: "1220000",
	}

	err := st.StorageMiners(m1)
	require.NoError(t, err)

	err := st.StorageMiners(m2)
	require.NoError(t, err)

	miner, err := st.MaxPowerMiner()
	require.NoError(t, err)
	t.log("miner: ", miner)
}

func TestMinPriceMiner(t *testing.T) {
	s, err := store.NewStorage("./storagehelper.db")
	require.NoError(t, err)

	ask1 := &api.Ask{
		Addr:   "t0001",
		AskID:  1,
		Price:  "100000",
		Expire: "255",
	}

	ask2 := &api.Ask{
		Addr:   "t0002",
		AskID:  1,
		Price:  "100002",
		Expire: "255",
	}

	err := s.StorageAsks(ask1)
	require.NoError(t, err)

	err := s.StorageAsks(ask2)
	require.NoError(t, err)

	ask, err := s.MinPriceMiner()
	require.NoError(t, err)

	t.log("ask :", ask)
}

func TestListMinerByPower(t *testing.T) {
	s, err := store.NewStorage("./storagehelper.db")
	require.NoError(t, err)

	miner, err := s.ListMinersbyPower(3)
	require.NoError(t, err)

	for _, m := range miner {
		t.log("miner:", m)
	}
}
