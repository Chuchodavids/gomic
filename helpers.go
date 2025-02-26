package main

import (
	"fmt"
	"os"
	"path"
)

func organizer(filename string, comicInfo ComicInfo) error {
	if comicInfo.Series != "" {
		p := path.Join(organizedPath, comicInfo.Series)
		err := os.MkdirAll(p, 0755)
		if err != nil {
			return fmt.Errorf("could not create path to organize the comics")
		}
		finalP := path.Join(p, comicInfo.Series+" #"+comicInfo.Number+getFileExtension(filename))
		err = os.Rename(filename, finalP)
		if err != nil {
			return fmt.Errorf("could not move file to the final destination")
		}
	}
	return nil
}
