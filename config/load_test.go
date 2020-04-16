package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToml(t *testing.T) {
	cfgP := "./config.toml"

	_, err := os.Stat(cfgP)
	require.NoError(t, err)

	c, err := os.Create(cfgP)
	require.NoError(t, err)

	comm, err := ConfigComment(DefaultStorageMiner())
	require.NoError(t, err)

	_, err = c.Write(comm)
	require.NoError(t, err)

	err = c.Close()
	require.NoError(t, err)

	var cfg StorageHelper
	_, err = FromFile("./config.toml", &cfg)
	require.NoError(t, err)
	t.Log("cfg: ", cfg)
}
