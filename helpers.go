package main

import (
	"os"
	"path"
	"regexp"
)

func extractTitleAndIssue(filename string) []string {
	re := regexp.MustCompile(`^(.*?) (\d+)\.(cbz|cbr|zip|rar|pdf)$`)
	match := re.FindStringSubmatch(filename)
	return match
}

func organizer(filename string, comicInfo ComicInfo) {
	if comicInfo.Series != "" {
		p := path.Join(organizedPath, comicInfo.Series)
		err := os.MkdirAll(p, 0755)
		if err != nil {
			panic(err)
		}
		finalP := path.Join(p, comicInfo.Series+" #"+comicInfo.Number+getFileExtension(filename))
		err = os.Rename(filename, finalP)
		if err != nil {
			panic(err)
		}
		return
	}

}
