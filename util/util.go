package util

import (
	"log"
	"os"
)

// PanicError panics whether error occures
func PanicError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// MkdirAll creates a directory named path, along with any neccessary parents
func MkdirAll(directory string) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, os.ModePerm)
	}
}
