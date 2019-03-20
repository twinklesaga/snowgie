package util

import (
	"io"
	"log"
	"net/http"
	"os"
)

type PathType int

const (
	PtUnknown PathType = iota
	PtFile
	PtDirectory
)

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}


func PathTypeCheck( path  string ) PathType{
	stat , err := os.Stat(path)

	if err != nil {
		return PtUnknown
	}
	if stat.IsDir() {
		return PtDirectory
	}
	return PtFile
}

func FileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	return fi.Size()
}

func DownloadFile( url string,filepath string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}