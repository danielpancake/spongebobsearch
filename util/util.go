package util

import (
	"encoding/json"
	"io/ioutil"
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

// JSONFromFile loads json from file and unmarshals to the interface given
func JSONFromFile(from string, to interface{}) {
	file, err := ioutil.ReadFile(from)
	PanicError(err)

	err = json.Unmarshal(file, to)
	PanicError(err)
}

func JSONToFile(from interface{}, to string) {
	file, err := os.Create(to)
	PanicError(err)

	json, err := json.Marshal(from)
	PanicError(err)

	file.Write(json)
	file.Close()
}
