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

	Pages []Page `xml:"Pages>Page"` //Nested struct
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
		comic.Summary = d.Paragraphs[0].Text
	}

	for _, creditPerson := range c.Credits {
		if comic.Writer != "" && comic.Penciller != "" && comic.Editor != "" {
			break
		}
		if creditPerson.Role == "writer" {
			comic.Writer = creditPerson.Name
			continue
		}
		if creditPerson.Role == "penciler" {
			comic.Penciller = creditPerson.Name
			continue
		}
		if creditPerson.Role == "editor" {
			comic.Editor = creditPerson.Name
		}
	}
	return comic, nil
}

func addComicInfoXml(comicFile string, comicInfo ComicInfo) error {
// Open the .cbz file for appending (or create it if it doesn't exist)
	cbzFile, err := os.OpenFile(comicFile, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer cbzFile.Close()

	tempComicInfo, _ := os.CreateTemp("", "comicinfo-*.xml")
	defer tempComicInfo.Close()

	file, err := xml.MarshalIndent(comicInfo, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(file))

	tempComicInfo.Write(file)

	// Create a new zip writer
	zipWriter := zip.NewWriter(cbzFile)
	defer zipWriter.Close()
	fileInfo, _ := os.Stat(tempComicInfo.Name())
	header, _ := zip.FileInfoHeader(fileInfo)
	header.Name = "comicinfo.xml"
	header.Method = zip.Deflate
	writer, _ := zipWriter.CreateHeader(header)
	// Create a new zip entry (file) inside the .cbz file

	// zipFile, err := zipWriter.Create("comicinfo.xml")

	// Copy the content of the file you want to add into the zip entry
	_, err = io.Copy(writer, tempComicInfo)
	if err != nil {
		return err
	}

	fmt.Println("File added successfully!")
	return nil
}

// func modifyComicInfo(){
// 	// Example modifications:
// 	comicInfo.Writer = "New Writer"
// 	comicInfo.Summary = "This is an updated summary."
// 	comicInfo.Genre = comicInfo.Genre + ", Comedy" // Append to existing genre

// 	//example adding a page
//     newPage := Page{Image: "001.jpg", Type: "FrontCover"}
//     comicInfo.Pages = append(comicInfo.Pages, newPage)

// 	// 5. Marshal (serialize) the modified data back to XML.
// 	modifiedXML, err := xml.MarshalIndent(comicInfo, "", "  ")
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal modified ComicInfo.xml: %w", err)
// 	}
//     modifiedXML = []byte(xml.Header + string(modifiedXML)) // Add XML header

// 	// 6. Create a temporary CBZ file.
// 	tempCBZPath := cbzPath + ".tmp" // Or use a more robust temp file method
// 	tempFile, err := os.Create(tempCBZPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create temporary CBZ file: %w", err)
// 	}
// 	defer tempFile.Close() // Make sure to close the temp file
// 	zipWriter := zip.NewWriter(tempFile)

// 	// 7. Write files to the temporary CBZ, replacing ComicInfo.xml.
// 	for _, f := range r.File {
// 		header, err := zip.FileInfoHeader(f.FileInfo())
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to create file header: %w", err)
// 		}

// 		//Crucially, set the correct name in the header:
// 		header.Name = f.Name
// 		header.Method = f.Method // Preserve compression method

// 		writer, err := zipWriter.CreateHeader(header)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to create file in zip: %w", err)
// 		}

// 		if f.Name == "ComicInfo.xml" {
// 			// Write the modified XML.
// 			if _, err := writer.Write(modifiedXML); err != nil {
// 				return nil, fmt.Errorf("failed to write modified ComicInfo.xml to zip: %w", err)
// 			}
// 		} else {
// 			// Copy other files from the original CBZ.
// 			reader, err := f.Open()
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to open file in original CBZ: %w", err)
// 			}
// 			if _, err := io.Copy(writer, reader); err != nil {
// 				reader.Close() //Close before returning.
// 				return nil, fmt.Errorf("failed to copy file to zip: %w", err)
// 			}
// 			reader.Close()
// 		}
// 	}

// 	if err := zipWriter.Close(); err != nil {
// 		return nil, fmt.Errorf("failed to close zip writer: %w", err)
// 	}

// 	// 8. Rename the temporary file to replace the original CBZ.
// 	if err := os.Rename(tempCBZPath, cbzPath); err != nil {
// 		//Attempt to remove temp file, even if rename fails.
// 		os.Remove(tempCBZPath)
// 		return nil, fmt.Errorf("failed to rename temporary CBZ file: %w", err)
// 	}
// }
