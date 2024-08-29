package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	sourceFolder = flag.String("src", "", "source folder")
	destinationFile = flag.String("o", "", "destination file")
)

// need to run from root of repo
func main() {
	flag.Parse()

	src, dst := *sourceFolder, *destinationFile
	fmt.Println("creating zip for: " + src)
	err := recursiveZip(src, dst)
	if err != nil {
		panic(err)
	}
	//	check file length for the zip
	f, err := os.Stat(dst)
	if err != nil {
		panic(err)
	}
	fmt.Println("file size: ", f.Size())
}

func recursiveZip(pathToZip, destinationPath string) error {
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	myZip := zip.NewWriter(destinationFile)
	err = filepath.Walk(pathToZip, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(filePath, filepath.Dir(pathToZip))
		zipFile, err := myZip.Create(relPath)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = myZip.Close()
	if err != nil {
		return err
	}
	return nil
}
