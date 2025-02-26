package main

import (
	"os"
	"path/filepath"
	"strings"
)

func run(comic string) error {

	// extract comic info from comicinfo.xml
	comicInfo, err := extractComicInfo(comic)
	if err != nil && err != ErrMissingComicInfo {
		return err
	}

	// if comicinfo.xml does not exists find it
	if err == ErrMissingComicInfo {
		// search in comic vine using name and issue number
		comic, _ = strings.CutSuffix(comic, filepath.Ext(comic))
		r, err := cvSearch(filepath.Base(comic))
		if err != nil {
			panic(err)
		}
		// Make another query to comics vine to get the credits from the comic ID
		credits, err := cvGetCredits(r.ID)
		if err != nil {
			panic(err)
		}
		// Add the credits returned array to the comic CV
		r.Credits = credits

		// Create the comicinfo.xml
		a, _ := createComicInfo(*r)

		addComicInfoXml(os.Args[1], a)

	}
	// if name and series in xml file, continue
	// TODO: fix this
	if err != ErrMissingComicInfo {
		organizer(comic, comicInfo)
	}
	return nil
}
