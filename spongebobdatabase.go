package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"spongebobdatabase/grabber"
	"spongebobdatabase/indexer"
	"spongebobdatabase/parser"
	"spongebobdatabase/util"
	"strings"

	"spongebobdatabase/types"

	"github.com/gookit/color"
)

func build() {
	database := grabber.GetContents()
	grabber.GetAllTranscripts(database)
	grabber.WriteAllTranscripts(database)

	contents := indexer.GenerateContents(database)
	util.CompressAndWrite(contents, "contents.gz")

	characters := indexer.GetAllCharacters(contents)
	util.CompressAndWrite(characters, "characters.gz")

	index := make(indexer.Index)
	index.AddAllToIndex(contents, characters)
	index.Build("index.gz")
}

func main() {
	filename := filepath.Base(os.Args[0])
	info := fmt.Sprintf(`usage: %s [-c | create] [-r | rebuild] [-s <query> | search " query "]
	
For the first initialization it is advised to run: %s -c
This command will grab SpongeBob transcripts and create both table of contents and index`, filename, filename)

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(info)
	} else {
		var query string
		c, r, s := false, false, false

		for _, arg := range args {
			switch arg {
			case "-c", "create":
				c = true
			case "-r", "rebuild":
				r = true
			case "-s", "search":
				s = true
			default:
				if s {
					query = arg
					s = false
				} else {
					fmt.Println("Invalid or unknown argument: " + arg)
					return
				}
			}
		}

		_, err := os.Stat("contents.gz")
		contentIsPresent := err == nil

		_, err = os.Stat("characters.gz")
		content2IsPresent := err == nil

		_, err = os.Stat("index.gz")
		indexIsPresent := err == nil

		if r {
			build()
		} else if c {
			if contentIsPresent || content2IsPresent || indexIsPresent {
				fmt.Println("Table of contents or index already exists! Run: " + filename + " -r to rebuild.")
				return
			}

			build()
		}

		if query != "" {
			if contentIsPresent && content2IsPresent && indexIsPresent {
				var contents types.Contents
				util.DecompressAndRead("contents.gz", &contents)

				var characters types.Contents
				util.DecompressAndRead("characters.gz", &characters)

				index := make(indexer.Index)
				index.Load("index.gz")

				results := index.Search(query)
				matches := len(results)
				if matches > 0 {
					fmt.Printf("\"%s\" has been found %d times:\n", query, matches)
				} else {
					fmt.Printf("No matches for \"%s\"", query)
				}

				for i, result := range results {
					var speaker string
					line, at := index.GetFromIndex(contents, characters, result)

					character := characters[result[1]]
					if character != "" {
						speaker = character + ": "
					}

					for _, q := range strings.Split(query, " ") {
						line = parser.Highlight(line, q, color.FgLightYellow)
					}

					fmt.Println(at)
					fmt.Println("\t" + speaker + line + "\n")

					if (i+1)%5 == 0 {
						fmt.Printf("Press 'Enter' to show more... (%d / %d)", i+1, matches)
						bufio.NewReader(os.Stdin).ReadBytes('\n')
						fmt.Println("")
					}
				}
			} else {
				fmt.Printf("Index is not found or has not been created yet! Please, run %s -c or %s -r before using search", filename, filename)
			}
		}
	}
}
