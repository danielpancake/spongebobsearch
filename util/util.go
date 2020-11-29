package util

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

// PanicError panics whenever error occures.
func PanicError(err error) {
	if err != nil {
		log.Printf("%s", err)
	}
}

// MkdirAll creates a directory named path, along with any neccessary parents.
func MkdirAll(directory string) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, os.ModePerm)
	}
}

// ValidateFilename removes all illegal characters from filename.
func ValidateFilename(text string) string {
	re, err := regexp.Compile("[^A-Za-z0-9-,&'\" ]+")
	PanicError(err)

	return re.ReplaceAllString(text, "")
}

// RemoveDuplicates removes duplicates in string slice.
func RemoveDuplicates(slice []string) []string {
	var output []string
	unique := make(map[string]bool)
	for _, entry := range slice {
		if _, value := unique[entry]; !value {
			output = append(output, entry)
			unique[entry] = true
		}
	}
	return output
}

// GetDocument returnes goquery.Document if no error occures.
func GetDocument(url string) *goquery.Document {
	page, err := http.Get(url)
	PanicError(err)

	defer page.Body.Close()
	if page.StatusCode != 200 {
		log.Fatalf("Error occured! Status Code: %d %s", page.StatusCode, page.Status)
	}

	document, err := goquery.NewDocumentFromReader(page.Body)
	PanicError(err)

	return document
}

// CompressAndWrite writes interface to the file using gzip comression.
func CompressAndWrite(from interface{}, filepath string) {
	var gz bytes.Buffer
	file, err := os.Create(filepath)
	PanicError(err)

	json, err := json.Marshal(from)
	PanicError(err)

	zipper := gzip.NewWriter(&gz)
	zipper.Write(json)
	zipper.Close()

	file.Write(gz.Bytes())
	file.Close()
}

// DecompressAndRead reads file compressed with gzip and decompresses it to the interface.
func DecompressAndRead(filepath string, to interface{}) {
	r, err := ioutil.ReadFile(filepath)
	PanicError(err)

	gzip, err := gzip.NewReader(bytes.NewReader(r))
	PanicError(err)

	result, err := ioutil.ReadAll(gzip)
	PanicError(err)

	err = json.Unmarshal(result, to)
	PanicError(err)
}

// JSONToFile saves json from data stucture to the path given.
func JSONToFile(from interface{}, filepath string) {
	file, err := os.Create(filepath)
	PanicError(err)

	json, err := json.Marshal(from)
	PanicError(err)

	file.Write(json)
	file.Close()
}

// JSONFromFile loads json from file and unmarshals to the interface given.
func JSONFromFile(from string, to interface{}) {
	file, err := ioutil.ReadFile(from)
	PanicError(err)

	err = json.Unmarshal(file, to)
	PanicError(err)
}
