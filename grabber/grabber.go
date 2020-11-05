package grabber

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

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

func writeTranscript(e Episode, transcript string, bar *pb.ProgressBar) {
	if transcript != "" {
		util.MkdirAll("output/" + e.category)

		file, err := os.Create(fmt.Sprintf("output/%s[%s] %s.txt", e.category, e.id, e.title))
		util.PanicError(err)

		bar.Increment()

		file.WriteString(transcript)
		file.Close()
	}
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

// GetAllEpisodes gets all transcipts from SpongeBob wiki
func GetAllEpisodes() {
	var episodes []Episode

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

	util.MkdirAll("output")
	for _, e := range episodes {
		go writeTranscript(e, getTranscript(e.link), bar)
	}

	bar.Finish()
}
