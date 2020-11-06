package grabber

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"spongebobdatabase/util"

	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb"
)

// Episode stores information about episode
type Episode struct {
	id       string
	title    string
	category string
	link     string
}

func getDocument(url string) *goquery.Document {
	page, err := http.Get(url)
	util.PanicError(err)

	defer page.Body.Close()
	if page.StatusCode != 200 {
		log.Fatalf("Error occured! Status Code: %d %s", page.StatusCode, page.Status)
	}

	document, err := goquery.NewDocumentFromReader(page.Body)
	util.PanicError(err)

	return document
}

func getTranscript(url string) string {
	transcript := getDocument("https://spongebob.fandom.com/" + url)
	return transcript.Find(".mw-parser-output ul").Text()
}

func writeTranscript(e Episode, transcript string) string {
	var filename string

	if transcript != "" {
		util.MkdirAll("output/" + e.category)

		filename = fmt.Sprintf("output/%s[%s] %s.txt", e.category, e.id, e.title)
		file, err := os.Create(filename)
		util.PanicError(err)

		json, err := json.Marshal(strings.Split(transcript, "\n"))
		util.PanicError(err)

		file.Write(json)
		file.Close()
	}

	return filename
}

func episodeExtractor(table *goquery.Selection, category string) []Episode {
	var temp []Episode
	table.Find("tbody tr").Each(func(i int, tr *goquery.Selection) {
		td := tr.Find("td").First()

		id := validateFilename(td.Text())
		title := validateFilename(td.Next().Text())
		link, exists := td.Next().Next().Find("a").Attr("href")

		if id != "" && title != "" && exists {
			temp = append(temp, Episode{id, title, category, link})
		}
	})
	return temp
}

func validateFilename(text string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9-,&'\" ]+")
	util.PanicError(err)

	return reg.ReplaceAllString(text, "")
}

// GetAllEpisodes gets all transcipts from SpongeBob wiki and returnes array or relative paths
func GetAllEpisodes() []string {
	var waiter sync.WaitGroup

	var episodes []Episode
	var index []string

	c := make(chan string)

	// Get information about all episodes
	var h2, h3, h4 string
	document := getDocument("https://spongebob.fandom.com/wiki/List_of_transcripts")

	document.Find(".mw-parser-output div:nth-child(2)").Children().Each(
		func(i int, header *goquery.Selection) {
			switch goquery.NodeName(header) {
			case "h2":
				h2 = validateFilename(header.Text()) + "/"
				h3 = ""
				h4 = ""

				table := header.Next()
				if table.Is(".wikitable") {
					episodes = append(episodes, episodeExtractor(table, h2)...)
				}
				break

			case "h3":
				h3 = validateFilename(header.Text()) + "/"
				h4 = ""

				table := header.Next()
				if table.Is(".wikitable") {
					episodes = append(episodes, episodeExtractor(table, h2+h3)...)
				}
				break

			case "h4":
				h4 = validateFilename(header.Text()) + "/"

				table := header.Next()
				if table.Is(".wikitable") {
					episodes = append(episodes, episodeExtractor(table, h2+h3+h4)...)
				}
				break
			}
		})

	bar := pb.StartNew(len(episodes))

	for _, e := range episodes {
		waiter.Add(1)

		go func(e Episode, bar *pb.ProgressBar) {
			transcript := getTranscript(e.link)

			if path := writeTranscript(e, transcript); path != "" {
				c <- path
			}

			bar.Increment()
			waiter.Done()
		}(e, bar)
	}

	// Wait until proccess is done
	go func() {
		waiter.Wait()
		bar.Finish()
		close(c)
	}()

	for path := range c {
		index = append(index, path)
	}

	return index
}
