package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var organizedPath string

func main() {
	organizedPath = "organized_comics"

	// parse the cbz file name from the argument
	if len(os.Args) != 2 {
		panic(fmt.Errorf("no file or directory defined"))
	}

	target, err := os.Stat(os.Args[1])
	if err != nil {
		panic(err)
	}

	if target.IsDir() {

		filepath.Walk(os.Args[1], func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			err = run(path)
			if err == ErrSkipIssue {
				return nil
			}
			if err != nil {
				return err
			}
			return nil
		})
	}
}
