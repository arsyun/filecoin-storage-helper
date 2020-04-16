package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path"
	"strings"
)

// Gzip and tar from source directory or file to destination file
// you need check file exist before you call this function
func TarGz(srcDirPath string, destFilePath string) error {
	fw, err := os.Create(destFilePath)
	if err != nil {
		return err
	}
	defer fw.Close()

	// Gzip writer
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// Tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Check if it's a file or a directory
	f, err := os.Open(srcDirPath)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.IsDir() {
		// handle source directory
		tarGzDir(srcDirPath, path.Base(srcDirPath), tw)
	} else {
		// handle file directly
		tarGzFile(srcDirPath, fi.Name(), tw, fi)
	}

	// log.Info("Well done!")
	return nil
}

// Deal with directories
// if find files, handle them with tarGzFile
// Every recurrence append the base path to the recPath
// recPath is the path inside of tar.gz
func tarGzDir(srcDirPath string, recPath string, tw *tar.Writer) error {
	// Open source diretory
	dir, err := os.Open(srcDirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	// Get file info slice
	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		// Append path
		curPath := srcDirPath + "/" + fi.Name()
		// Check it is directory or file
		if fi.IsDir() {
			// (Directory won't add unitl all subfiles are added)
			tarGzDir(curPath, recPath+"/"+fi.Name(), tw)
		}
		//  else {
		// File
		// log.Infof("Adding file...%s", curPath)
		// }

		tarGzFile(curPath, recPath+"/"+fi.Name(), tw, fi)
	}

	return nil
}

// Deal with files
func tarGzFile(srcFile string, recPath string, tw *tar.Writer, fi os.FileInfo) error {
	if fi.IsDir() {
		// Create tar header
		hdr := new(tar.Header)
		// if last character of header name is '/' it also can be directory
		// but if you don't set Typeflag, error will occur when you untargz
		hdr.Name = recPath + "/"
		hdr.Typeflag = tar.TypeDir
		hdr.Size = 0
		//hdr.Mode = 0755 | c_ISDIR
		hdr.Mode = int64(fi.Mode())
		hdr.ModTime = fi.ModTime()

		// Write hander
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
	} else {
		// File reader
		fr, err := os.Open(srcFile)
		if err != nil {
			return err
		}
		defer fr.Close()

		// Create tar header
		hdr := new(tar.Header)
		hdr.Name = recPath
		hdr.Size = fi.Size()
		hdr.Mode = int64(fi.Mode())
		hdr.ModTime = fi.ModTime()

		// Write hander
		if err = tw.WriteHeader(hdr); err != nil {
			return err
		}

		// Write file data
		if _, err = io.Copy(tw, fr); err != nil {
			return err
		}
	}

	return nil
}

// Ungzip and untar from source file to destination directory
// you need check file exist before you call this function
func UnTarGz(srcFilePath string, destDirPath string) (string, error) {
	// Create destination directory
	os.Mkdir(destDirPath, os.ModePerm)

	fr, err := os.Open(srcFilePath)
	if err != nil {
		// log.Errorf("untargz fail, open %s err: %+v", srcFilePath, err)
		return "", err
	}
	defer fr.Close()

	// Gzip reader
	gr, err := gzip.NewReader(fr)

	// Tar reader
	tr := tar.NewReader(gr)
	var dirName string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		// log.Infof("UnTarGzing file %s..." + hdr.Name)
		dirName = strings.Split(hdr.Name, "/")[0]
		// Check if it is diretory or file
		if hdr.Typeflag != tar.TypeDir {
			// Get files from archive
			// Create diretory before create file
			os.MkdirAll(destDirPath+"/"+path.Dir(hdr.Name), os.ModePerm)
			fw, _ := os.Create(destDirPath + "/" + hdr.Name)

			if _, err = io.Copy(fw, tr); err != nil {
				return "", err
			}
		}
	}

	// log.Info("Well done!")
	return dirName, nil
}
