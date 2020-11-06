package search

import (
	"os"
	"spongebobdatabase/util"
	"strings"
	"unicode"

	snowman "github.com/kljensen/snowball/english"
)

// Index stores indexed spongebobdatabase in memory
type Index map[string][]Location

// Location stores
type Location [2]int

var stopwords = map[string]struct{}{
	"a": {}, "and": {}, "be": {}, "have": {}, "i": {},
	"in": {}, "of": {}, "that": {}, "the": {}, "to": {},
}

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

func stopwordFilter(tokens []string) []string {
	r := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, exists := stopwords[token]; !exists {
			r = append(r, token)
		}
	}
	return r
}

func stemmerFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = snowman.Stem(token, false)
	}
	return r
}

func analyze(input string) []string {
	tokens := tokenize(input)
	tokens = lowercaseFilter(tokens)
	tokens = stopwordFilter(tokens)
	tokens = stemmerFilter(tokens)
	return tokens
}

// AddToIndex parses and adds index about given transript
func (index Index) AddToIndex(id int, path string) {
	var transcript []string
	util.JSONFromFile(path, &transcript)

	for i, sentence := range transcript {
		tokens := analyze(sentence)
		for _, token := range tokens {
			value := index[token]
			coord := Location{id, i}

			if value != nil && value[len(value)-1] == coord {
				continue
			}

			index[token] = append(value, coord)
		}
	}
}

// LoadFromFile loads index from file to memory
func (index Index) LoadFromFile(path string) {
	util.JSONFromFile(path, &index)
}

// Search searches through index and returns array of location of matches
func (index Index) Search(query string) []Location {
	var out []Location

	tokens := analyze(query)
	for _, token := range tokens {
		value := index[token]

		if value != nil {
			out = append(out, value...)
		}
	}

	return out
}

// GetFromIndex gets specified part of transcipt
func (index Index) GetFromIndex(contents []string, coord Location) string {
	var out string
	path := "output/" + contents[coord[0]]

	if _, err := os.Stat(path); err == nil {
		var data []string
		util.JSONFromFile(path, &data)
		out = data[coord[1]]
	}

	return out
}
