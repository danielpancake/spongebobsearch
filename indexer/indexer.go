package indexer

import (
	"fmt"
	"spongebobdatabase/grabber"
	"spongebobdatabase/parser"
	"spongebobdatabase/types"
	"spongebobdatabase/util"

	"github.com/cheggaaa/pb"
)

// Index stores map of words with their location.
type Index map[string][]Location

// Location stores three digits representing transcript index, character index and number of sentence.
type Location [3]int

// GenerateContents generates table of contents which contains of relative paths to the transcripts.
func GenerateContents(ts []grabber.TranscriptDS) types.Contents {
	var output types.Contents
	fmt.Println("Generating table of contents, please wait...")
	for _, t := range ts {
		output = append(output, fmt.Sprintf("%s[%s] %s", t.Category, t.ID, t.Title))
	}
	return output
}

// GetCharacters returns an array of characters from single transcript.
func GetCharacters(filepath string) types.Contents {
	var characters types.Contents
	var temp types.Transcript
	util.JSONFromFile("output/"+filepath+".txt", &temp)

	for c := range temp {
		characters = append(characters, c)
	}
	return characters
}

// GetAllCharacters returns an array of unique characters from every transcript.
func GetAllCharacters(contents types.Contents) types.Contents {
	var characters types.Contents
	fmt.Println("Generating table of characters, please wait...")
	for _, c := range contents {
		characters = append(characters, GetCharacters(c)...)
	}
	characters = util.RemoveDuplicates(characters)
	return characters
}

// GetCharacterID returns number representing character index.
func GetCharacterID(characters types.Contents, character string) int {
	for i := 0; i < len(characters); i++ {
		if characters[i] == character {
			return i
		}
	}
	return -1
}

// AddToIndex parses and adds transcript to the index.
func (index Index) AddToIndex(id int, characters types.Contents, filepath string) {
	var transcript types.Transcript
	util.JSONFromFile("output/"+filepath+".txt", &transcript)

	for character, lines := range transcript {
		characterID := GetCharacterID(characters, character)

		for i, line := range lines {
			tokens := parser.Analyze(line)

			for _, t := range tokens {
				value := index[t]
				coord := Location{id, characterID, i}

				if value != nil && value[len(value)-1] == coord {
					continue
				}

				index[t] = append(value, coord)
			}
		}
	}
}

// AddAllToIndex parses and adds all transcripts to the index.
func (index Index) AddAllToIndex(contents types.Contents, characters types.Contents) {
	fmt.Println("Building index, please wait...")
	bar := pb.StartNew(len(contents))
	for i, c := range contents {
		index.AddToIndex(i, characters, c)
		bar.Increment()
	}
	bar.Finish()
}

// Build copresses and saves index to file.
func (index Index) Build(filepath string) {
	util.CompressAndWrite(index, filepath)
	fmt.Println("\nDone! Now you're ready to search through SpongeBob database ;)")
}

// Load loads and decompresses index from index.gz file.
func (index Index) Load(filepath string) {
	util.DecompressAndRead(filepath, &index)
}

// Search does an index search and returns array of match locations.
func (index Index) Search(query string) []Location {
	var output []Location

	tokens := parser.Analyze(query)
	for _, token := range tokens {
		value := index[token]

		if value != nil {
			if output == nil {
				output = append(output, value...)
			} else {
				output = intersection(output, value)
			}
		}
	}

	return output
}

func intersection(a []Location, b []Location) []Location {
	var output []Location

	for _, aa := range a {
		for _, bb := range b {
			if aa == bb {
				output = append(output, aa)
			}
		}
	}

	return output
}

// GetFromIndex gets specified part of transcipt.
func (index Index) GetFromIndex(contents types.Contents, characters types.Contents, coord Location) (string, string) {
	filepath := "output/" + contents[coord[0]] + ".txt"
	character := characters[coord[1]]

	transcript := make(types.Transcript)
	util.JSONFromFile(filepath, &transcript)

	return transcript[character][coord[2]], contents[coord[0]]
}
