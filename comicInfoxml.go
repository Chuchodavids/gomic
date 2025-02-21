package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ComicInfo struct {
	XMLName     xml.Name `xml:"ComicInfo"`
	Title       string   `xml:"Title"`
	Series      string   `xml:"Series"`
	Number      string   `xml:"Number"`
	Editor      string   `xml:"Editor"`
	Summary     string   `xml:"Summary"`
	Writer      string   `xml:"Writer"`
	Penciller   string   `xml:"Penciller"`
	Genre       string   `xml:"Genre"`
	PageCount   int      `xml:"PageCount"` //Example of int
	LanguageISO string   `xml:"LanguageISO"`
	Publisher   string   `xml:"Publisher"`
	Year        int      `xml:"Year"`
	Month       int      `xml:"Month"`
	Day         int      `xml:"Day"`
	Inker       string   `xml:"Inker"`
	Letterer    string   `xml:"letterer"`
	Pages       []Page   `xml:"Pages>Page"` //Nested struct
	Colorist    string   `xml:"Colorist"`
	CoverArtist string `xml:"CoverArtist"`
}

type Page struct {
	Image       string `xml:"Image,attr"`
	Type        string `xml:"Type,attr,omitempty"` // Example: "FrontCover"
	DoublePage  bool   `xml:"DoublePage,attr,omitempty"`
	ImageSize   int    `xml:"ImageSize,attr,omitempty"`
	Key         string `xml:"Key,attr,omitempty"`
	ImageWidth  int    `xml:"ImageWidth,attr,omitempty"`
	ImageHeight int    `xml:"ImageHeight,attr,omitempty"`
}

type Paragraph struct {
	Text  string `xml:",chardata"`
	Links []Link `xml:"a"`
}

// Link structure represents <a> and its attributes
type Link struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
}

// Document structure to hold the parsed XML content
type Document struct {
	Headings   []string    `xml:"h4"` // Holds all <h4> headings
	Paragraphs []Paragraph `xml:"p"`  // Holds all <p> paragraphs
}

func parseXML(xmlContent []byte) (Document, error) {
	var doc Document
	decoder := xml.NewDecoder(strings.NewReader(string(xmlContent)))

	// Traverse through the XML structure
	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break // End of the document
			}
			return doc, err
		}

		switch el := token.(type) {
		case xml.StartElement:
			// Start of <h4> or <p> element
			if el.Name.Local == "h4" {
				var heading string
				decoder.DecodeElement(&heading, &el)
				doc.Headings = append(doc.Headings, heading)
			}
			if el.Name.Local == "p" {
				var para Paragraph
				decoder.DecodeElement(&para, &el)
				doc.Paragraphs = append(doc.Paragraphs, para)
			}
		}
	}

	return doc, nil
}
func extractComicInfo(cbzPath string) (ComicInfo, error) {
	// 1. Open the CBZ file.
	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		return ComicInfo{}, fmt.Errorf("failed to open CBZ file: %w", err)
	}
	defer r.Close()

	var comicInfoXML []byte

	// 2. Find and extract ComicInfo.xml.
	for _, f := range r.File {
		if f.Name == "ComicInfo.xml" {
			rc, err := f.Open()
			if err != nil {
				return ComicInfo{}, fmt.Errorf("failed to open ComicInfo.xml within CBZ: %w", err)
			}
			comicInfoXML, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return ComicInfo{}, fmt.Errorf("failed to read ComicInfo.xml: %w", err)
			}
			break // Exit loop once found
		}
	}

	if comicInfoXML == nil {
		return ComicInfo{}, ErrMissingComicInfo
	}

	// 3. Unmarshal (parse) the XML.
	var comicInfo ComicInfo
	if err := xml.Unmarshal(comicInfoXML, &comicInfo); err != nil {
		return ComicInfo{}, fmt.Errorf("failed to unmarshal ComicInfo.xml: %w", err)
	}

	return comicInfo, nil
}

// Helper function to get file extension
func getFileExtension(filePath string) string {
	return filepath.Ext(filePath)
}

func createComicInfo(c CVResult) (ComicInfo, error) {
	d, err := parseXML([]byte(c.Description))
	if err != nil {
		return ComicInfo{}, err
	}
	comic := ComicInfo{
		Title:  c.Name,
		Series: c.Volume.Name,
		Number: c.IssueNumber,
	}
	if len(d.Paragraphs) != 0 {
		var summary []string
		for _, par := range d.Paragraphs {
			summary = append(summary, par.Text)
		}
		comic.Summary = strings.Join(summary, "\n")
		comic.Summary = strings.TrimSpace(comic.Summary)
	}

	var colorist []string
	var writer []string
	var penciler []string
	var letterer []string
	var inker []string
	var editor []string
	var coverArtist []string
	for _, creditPerson := range c.Credits {
		switch creditPerson.Role {
		case "writer":
			writer = append(writer, creditPerson.Name)
		case "penciler":
			penciler = append(penciler, creditPerson.Name)
		case "editor":
			editor = append(editor, creditPerson.Name)
		case "letterer":
			letterer = append(letterer, creditPerson.Name)
		case "inker":
			inker = append(inker, creditPerson.Name)
		case "colorist":
			colorist = append(colorist, creditPerson.Name)
		case "cover":
			coverArtist = append(coverArtist, creditPerson.Name)
		}
	}
	comic.Writer = strings.Join(writer, ",")
	comic.Editor = strings.Join(editor, ",")
	comic.Colorist = strings.Join(colorist, ",")
	comic.Penciller = strings.Join(penciler, ",")
	comic.Letterer = strings.Join(letterer, ",")
	comic.Inker = strings.Join(inker, ",")
	comic.CoverArtist = strings.Join(coverArtist, ",")

	return comic, nil
}

func addComicInfoXml(comicFile string, comicInfo ComicInfo) error {
	// Open the .cbz file for appending (or create it if it doesn't exist)
	cbzFile, err := os.OpenFile(comicFile, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer cbzFile.Close()

	tempCBZfile, err := os.CreateTemp("", "temp-cbz-*.cbz")
	if err != nil {
		return err
	}
	defer tempCBZfile.Close()

	zipWriter := zip.NewWriter(tempCBZfile)
	defer zipWriter.Close()

	fileinfo, _ := cbzFile.Stat()

	zipReader, err := zip.NewReader(cbzFile, fileinfo.Size())
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		zipFile, err := zipWriter.Create(file.Name)
		if err != nil {
			return err
		}

		_, err = io.Copy(zipFile, fileReader)
		if err != nil {
			return err
		}
	}

	comicInfoXml, err := xml.MarshalIndent(comicInfo, "", "  ")
	if err != nil {
		return err
	}

	zipFile, _ := zipWriter.Create("ComicInfo.xml")

	_, err = zipFile.Write(comicInfoXml)
	if err != nil {
		return err
	}

	if err := os.Rename(tempCBZfile.Name(), comicFile); err != nil {
		return err
	}

	return nil
}
