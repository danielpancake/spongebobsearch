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

// TranscriptDS is a data structure which stores transcript along with its formation.
type TranscriptDS struct {
	ID         string
	Title      string
	Category   string
	URL        string
	Transcript types.Transcript
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
				temp = append(temp, TranscriptDS{id, title, category, url, nil})
			}
		})

	return temp
}

// GetTranscript gets transcript as text.
func (t *TranscriptDS) GetTranscript() {
	var temp string

	util.GetDocument(base + t.URL).Find(".mw-parser-output ul li").Each(
		func(i int, s *goquery.Selection) {
			temp += s.Text() + "\n"
		})

	t.Transcript = parser.ParseTranscript(temp)
}

// GetAllTranscripts gets all transcripts.
func GetAllTranscripts(ts []TranscriptDS) {
	var waiter sync.WaitGroup
	fmt.Println("Scraping transripts, please wait...")
	bar := pb.StartNew(len(ts))

	for i := 0; i < len(ts); i++ {
		waiter.Add(1)

		go func(t *TranscriptDS, bar *pb.ProgressBar) {
			t.GetTranscript()
			bar.Increment()
			waiter.Done()
		}(&ts[i], bar)
	}

	waiter.Wait()
	bar.Finish()
}

// WriteTranscript saves transcript to the file
func (t *TranscriptDS) WriteTranscript() {
	util.MkdirAll("output/" + t.Category)
	util.JSONToFile(t.Transcript, fmt.Sprintf("output/%s[%s] %s.txt", t.Category, t.ID, t.Title))
}

// WriteAllTranscripts writes all transcripts.
func WriteAllTranscripts(ts []TranscriptDS) {
	var waiter sync.WaitGroup
	fmt.Println("Parsing and writing transripts, please wait...")
	bar := pb.StartNew(len(ts))

	for i := 0; i < len(ts); i++ {
		waiter.Add(1)

		go func(t *TranscriptDS, bar *pb.ProgressBar) {
			t.WriteTranscript()
			waiter.Done()
			bar.Increment()
		}(&ts[i], bar)
	}

	waiter.Wait()
	bar.Finish()
}
