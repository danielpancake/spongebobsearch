package indexer

import (
	"encoding/json"
	"io/ioutil"
	"spongebobdatabase/util"
	"strings"
	"unicode"

	snowman "github.com/kljensen/snowball/english"
)

// Index stores indexed spongebobdatabase in memory
type Index map[string][][2]int

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

	file, err := ioutil.ReadFile(path)
	util.PanicError(err)

	json.Unmarshal(file, &transcript)

	for i, sentence := range transcript {
		tokens := analyze(sentence)
		for _, token := range tokens {
			value := index[token]
			location := [2]int{id, i}

			if value != nil && value[len(value)-1] == location {
				continue
			}

			index[token] = append(value, location)
		}
	}
}
