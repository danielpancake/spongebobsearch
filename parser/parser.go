package parser

import (
	"regexp"
	"spongebobdatabase/types"
	"spongebobdatabase/util"
	"strings"
	"unicode"

	"github.com/dlclark/regexp2"
	snowball "github.com/kljensen/snowball/english"
)

// List of ignored words
var stopwords = map[string]struct{}{
	"a": {}, "the": {}, "i": {}, "am": {}, "you": {},
	"are": {}, "is": {}, "be": {}, "have": {}, "to": {},
	"on": {},
}

// ParseTranscript parses transcript.
func ParseTranscript(transcript string) types.Transcript {
	output := make(types.Transcript)

	re, err := regexp.Compile("^[^[\\]():]+:")
	util.PanicError(err)

	re2, err := regexp2.Compile("(?:(\\.{3}([a-z]|[ ]*\\[)))|(?:(?<=\\[)[^]]*(?=\\]))|(?<![^.])]|(?<!Mr|Mrs)[?!.](?=[A-Z[]|$|\n| [A-Z[])", 0)
	util.PanicError(err)

	paragraphs := strings.Split(transcript, "\n")
	for _, paragraph := range paragraphs {
		character := strings.Trim(re.FindString(paragraph), "\u00A0: ")

		if l := len(character); l > 0 {
			paragraph = paragraph[l+1:]
		}

		temp := output[character]

		s, err := re2.ReplaceFunc(paragraph,
			func(m regexp2.Match) string {
				if m.Length > 1 {
					return m.String()
				}
				return m.String() + "\n"
			}, -1, -1)
		util.PanicError(err)

		lines := strings.Split(s, "\n")
		lastline := ""
		for _, line := range lines {
			if line = strings.Trim(line, " "); line != "" {

				// Concatenate short phrases:
				if len(lastline+line) > 60 {
					temp = append(temp, lastline+line)
					lastline = ""
				} else {
					lastline += line + " "
				}
			}
		}

		if lastline != "" {
			temp = append(temp, strings.Trim(lastline, " "))
		}
		if temp != nil {
			output[character] = util.RemoveDuplicates(temp)
		}
	}
	return output
}

// Tokenize returns an array of separated words.
func Tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '\''
	})
}

// FilterLowercase puts all strings in the array to their lowercase.
func FilterLowercase(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

// FilterStopwords removes stopwords from the array.
func FilterStopwords(tokens []string) []string {
	var r []string
	for _, token := range tokens {
		if _, exists := stopwords[token]; !exists {
			r = append(r, token)
		}
	}
	return r
}

// FilterStemmer returns an array of word stems.
func FilterStemmer(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = snowball.Stem(token, false)
	}
	return r
}

// Analyze puts words to their lowercase, removes stopwords words and returns an array of word stems.
func Analyze(input string) []string {
	tokens := Tokenize(input)
	tokens = FilterLowercase(tokens)
	tokens = FilterStopwords(tokens)
	tokens = FilterStemmer(tokens)
	return tokens
}
