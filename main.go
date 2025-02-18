package main

import (
	"fmt"
	"os"
)

var organizedPath string

func main() {
	organizedPath = "organized_comics"

	// parse the file
	arg := os.Args[1:]
	if len(arg) != 1 {
		panic(fmt.Errorf("no file defined"))
	}
	// read if the file exists
	file, err := os.Stat(arg[0])
	if err != nil {
		panic(err)
	}
	// extrac the information from comicinfo.xml
	comicInfo, err := extractComicInfo(arg[0])
	if err != nil && err != ErrMissingComicInfo {
		panic(err)
	}

	// if name and series in xml file, continue
	// TODO: fix this
	if err != ErrMissingComicInfo {
		organizer(arg[0], comicInfo)
	}

	// if no xml info then use comicvine
	heuName := extractTitleAndIssue(file.Name())

	// search in comic vine using name and issue number
	r, err := cvSearch(heuName[1], heuName[2])
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
