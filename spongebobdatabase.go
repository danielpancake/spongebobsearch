package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"spongebobdatabase/grabber"
	"spongebobdatabase/search"
	"spongebobdatabase/util"
)

func build() {
	fmt.Println("Getting all SpongeBob transcripts, please wait...")
	contents := grabber.GetAllEpisodes()
	util.JSONToFile(contents, "contents.txt")

	fmt.Print("\nBuilding index, please wait... ")
	index := make(search.Index)
	for id, path := range contents {
		if _, err := os.Stat("output/" + path); err == nil {
			index.AddToIndex(id, "output/"+path)
		}
	}
	fmt.Println("Done!")

	util.JSONToFile(index, "index.txt")
}

func main() {
	filename := filepath.Base(os.Args[0])
	info := fmt.Sprintf(`usage: %s [-c | create] [-r | rebuild] [-s <query> | search " query "]
	
For the first initialisation it is advised to run: %s create
This command will grab SpongeBob transcripts and create both table of contents and index`, filename, filename)

	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println(info)
	} else {
		var query string

		c := false
		r := false
		s := false

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

		_, err := os.Stat("contents.txt")
		contentIsPresent := err == nil

		_, err = os.Stat("index.txt")
		indexIsPresent := err == nil

		if r {
			build()
		} else if c {
			if contentIsPresent {
				fmt.Println("Table of contents already exists! Run: " + filename + " rebuild")
				if indexIsPresent {
					fmt.Println("Getting all SpongeBob transcripts, please wait...")
				}
			} else {
				build()
			}
		}

		if query != "" {
			if contentIsPresent && indexIsPresent {
				var contents []string
				util.JSONFromFile("contents.txt", &contents)

				index := make(search.Index)

				index.LoadFromFile("index.txt")
				results := index.Search(query)

				num := len(results)
				if num > 0 {
					fmt.Printf("Word \"%s\" has been found %d times:\n", query, num)
				} else {
					fmt.Printf("No matches for word \"%s\"", query)
				}

				for i, result := range results {
					paragraph, at := index.GetFromIndex(contents, result)

					fmt.Println(at)
					fmt.Println(`	` + paragraph + "\n")

					if (i+1)%5 == 0 {
						fmt.Printf("Press 'Enter' to show more... (%d / %d)", i+1, num)
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
