// This tool used to prepare data for database migration and initialize data.
// It copies folder with migration scripts and file prod_data.sql from package
// data to package statik.

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

func main() {
	// copies migration scripts
	err := copyDir("../data/migration/scripts", "./scripts/migration")
	if err != nil {
		panic(fmt.Errorf("failed when copy dir %v", err))
	}

	// copy initialize data script
	err = copyFile("../data/prod_data.sql", "./scripts/prod_data.sql")
	if err != nil {
		panic(fmt.Errorf("failed when copy file %v", err))
	}
}

func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func copyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			err = copyDir(srcfp, dstfp)
		} else {
			err = copyFile(srcfp, dstfp)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
