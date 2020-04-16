package utils

import (
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("utils")

func GenerateDBName(filepath string) string {
	new := strings.Replace(filepath, "/", "_", -1)
	return strings.TrimPrefix(new, "_")
}

func ComputeChunks(totalsize uint64, slicesize uint64) int {
	num := int(math.Ceil(float64(totalsize) / float64(slicesize)))
	return num
}

func FileChecker(filepath string) (*os.File, error) {
	f, err := os.Open(filepath)
	if err != nil && os.IsNotExist(err) {
		tf, err := os.Create(filepath)
		if err != nil {
			return nil, err
		}
		return tf, err
	}

	return f, nil
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}

	return true
}

func GenerateFileByPath(fullpath string) error {
	if Exists(fullpath) {
		return nil
	}
	dirs := path.Dir(fullpath)
	if err := os.MkdirAll(dirs, 0755); err != nil {
		return err
	}

	_, err := os.Create(fullpath)
	if err != nil {
		return err
	}

	return nil
}

func RemoveFileOrDir(path string) error {
	if !Exists(path) {
		return nil
	}
	f, err := os.Stat(path)
	if err != nil {
		return err
	}

	if f.IsDir() {
		os.RemoveAll(path)
	} else {
		os.Remove(path)
	}

	return nil
}

func GetToken() (string, error) {
	tokenpath := "/root/.lotus/token"
	data, err := ioutil.ReadFile(tokenpath)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(data), "\n"), nil
}
