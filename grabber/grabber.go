package grabber

import (
	"fmt"
	"spongebobdatabase/parser"
	"spongebobdatabase/types"
	"spongebobdatabase/util"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb"
)

// TranscriptDS is a data structure which stores information about transcript.
type TranscriptDS struct {
	ID       string
	Title    string
	Category string
	URL      string
}

var base string = "https://spongebob.fandom.com/"

// GetContents gets information about every episode, short or movie from SpongeBob wiki page.
// It returns array of transcript's information.
func GetContents() []TranscriptDS {
	var contents []TranscriptDS
	fmt.Println("Collecting information from SpongeBob wiki page, please wait...")

	var h2, h3, h4 string
	wikipage := util.GetDocument(base + "/wiki/List_of_transcripts")
	wikipage.Find(".mw-parser-output div:nth-child(2)").Children().Each(
		func(i int, section *goquery.Selection) {
			switch goquery.NodeName(section) {
			case "h2":
				h2 = util.ValidateFilename(section.Text()) + "/"
				h3 = ""
				h4 = ""

				if table := section.Next(); table.Is(".wikitable") {
					contents = append(contents, contentExtractor(table, h2)...)
				}
			case "h3":
				h3 = util.ValidateFilename(section.Text()) + "/"
				h4 = ""

				if table := section.Next(); table.Is(".wikitable") {
					contents = append(contents, contentExtractor(table, h2+h3)...)
				}
			case "h4":
				h4 = util.ValidateFilename(section.Text()) + "/"

				if table := section.Next(); table.Is(".wikitable") {
					contents = append(contents, contentExtractor(table, h2+h3+h4)...)
				}
			}
		})

	return contents
}

func contentExtractor(section *goquery.Selection, category string) []TranscriptDS {
	var temp []TranscriptDS

	section.Find("tbody tr").Each(
		func(i int, tr *goquery.Selection) {
			td := tr.Find("td").First()

			id := util.ValidateFilename(td.Text())
			title := util.ValidateFilename(td.Next().Text())
			url, exists := td.Next().Next().Find("a").Attr("href")

			if id != "" && title != "" && exists {
				temp = append(temp, TranscriptDS{id, title, category, url})
			}
		})

	return temp
}

// GetTranscript returns transcript.
func GetTranscript(t TranscriptDS) types.Transcript {
	var temp string
	util.GetDocument(base + t.URL).Find(".mw-parser-output ul li").Each(
		func(i int, s *goquery.Selection) {
			temp += s.Text() + "\n"
		})
	return parser.ParseTranscript(temp)
}

// WriteTranscript writes transcript to the file.
func WriteTranscript(t TranscriptDS, transcript types.Transcript) {
	util.MkdirAll("output/" + t.Category)
	util.JSONToFile(transcript, fmt.Sprintf("output/%s[%s] %s.txt", t.Category, t.ID, t.Title))
}

// GrabTranscript gets transcript and writes it to the file.
func GrabTranscript(t TranscriptDS) {
	WriteTranscript(t, GetTranscript(t))
}

// GrabAllTranscripts gets and writes all transcripts to the files.
func GrabAllTranscripts(ts []TranscriptDS, goroutines bool) {
	var waiter sync.WaitGroup
	fmt.Println("Scraping, parsing and writing transripts, please wait...")
	bar := pb.StartNew(len(ts))

	grab := func(t TranscriptDS, bar *pb.ProgressBar) {
		GrabTranscript(t)
		bar.Increment()
		waiter.Done()
	}

	for i := 0; i < len(ts); i++ {
		waiter.Add(1)
		if goroutines {
			go grab(ts[i], bar)
		} else {
			grab(ts[i], bar)
		}
	}

	waiter.Wait()
	bar.Finish()
}
