package utils_test

import (
	"go-filecoin-storage-helper/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTarGz(t *testing.T) {
	srcdirpath := "/home/john/testtargz/1.txt"

	err := utils.GenerateFileByPath(srcdirpath)
	require.NoError(t, err)

	utils.TarGz(srcdirpath, "/home/john/testtargz/1.tar.gz")
}

func TestUnTarGz(t *testing.T) {
	targzpath := "/home/john/testtargz/1.tar.gz"
	destdirpath := "/home/john/testtargz"

	_, err := utils.UnTarGz(targzpath, destdirpath)
	require.NoError(t, err)
}
