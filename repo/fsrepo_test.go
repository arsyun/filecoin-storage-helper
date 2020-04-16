package repo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepoInit(t *testing.T) {
	repopath := "/root/.storagehelper"
	r, err := NewFS(repopath)
	require.NoError(t, err)

	err = r.Init()
	require.NoError(t, err)
}
