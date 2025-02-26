package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func run(comicPath string) error {

	// extract comic info from comicinfo.xml
	comicInfo, err := extractComicInfo(comicPath)
	if err != nil && err != ErrMissingComicInfo {
		return err
	}

	// if comicinfo.xml does not exists find it
	if err == ErrMissingComicInfo {
		// search in comic vine using name and issue number
		comic, ok := strings.CutSuffix(comicPath, filepath.Ext(comicPath))
		if !ok {
			return fmt.Errorf("file has no extension")
		}
		comicFilename := filepath.Base(comic)
		r, err := cvSearch(strings.ToLower(comicFilename))
		if err != nil {
			return err
		}
		// Make another query to comics vine to get the credits from the comic ID
		credits, err := cvGetCredits(r.ID)
		if err != nil {
			panic(err)
		}
		// Add the credits returned array to the comic CV
		r.Credits = credits

		// Create the comicinfo.xml
		newComicInfoXml, err := createComicInfo(*r)
		if err != nil {
			return err
		}

		err = addComicInfoXml(comicPath, newComicInfoXml)
		if err != nil {
			return err
		}

	}
	// if name and series in xml file, continue
	// TODO: fix this
		err = organizer(comicPath, comicInfo)
		if err != nil {
			return err
		}
	return nil
}
