package web

import (
	"archiscribe/lib"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
)

func pickVolume(year int) string {
	for {
		entry := lib.IDCache.Random(year)
		candidate := entry.Identifier
		isFrak, _ := lib.IsFraktur(candidate)
		if !isFrak {
			log.Printf("%s is not fraktur, continuing search\n", candidate)
			continue
		}
		return candidate
	}
}

type lineProducer struct {
	resp     http.ResponseWriter
	ident    string
	year     int
	taskSize int
	progChan chan lib.ProgressMessage
	lineChan chan []lib.OCRLine
}

func newLineProducer(resp http.ResponseWriter, taskSize int, year int) (*lineProducer, error) {
	if _, ok := resp.(http.Flusher); !ok {
		return nil, fmt.Errorf("streaming unsupported")
	}
	if _, ok := resp.(http.CloseNotifier); !ok {
		return nil, fmt.Errorf("close notification unsupported")
	}
	if taskSize == 0 {
		taskSize = 50
	}
	return &lineProducer{resp: resp, taskSize: taskSize, year: year}, nil
}

func (p *lineProducer) produceLines() {
	p.ident = pickVolume(p.year)
	p.progChan, p.lineChan = lib.FetchLines(p.ident)
	log.Printf("Getting lines for %s", p.ident)
	headers := p.resp.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")

	metadata, _ := lib.GetMetadata(p.ident)
	doc := lib.Document{
		Identifier: p.ident,
		Title:      metadata.Get("title").MustString(),
		Year:       p.year,
		Manifest:   fmt.Sprintf("https://iiif.archivelab.org/iiif/%s/manifest.json", p.ident),
	}
	p.writeMessage("document", doc)
	p.streamLines()
}

func (p *lineProducer) writeMessage(event string, msg interface{}) {
	json, _ := json.Marshal(msg)
	fmt.Fprintf(p.resp, "event: %s\n", event)
	fmt.Fprintf(p.resp, "data: %s\n\n", json)
	p.resp.(http.Flusher).Flush()
}

func (p *lineProducer) handleLines(lines []lib.OCRLine) {
	lineIdxes := make([]int, 0, p.taskSize)
	lineIdxesMap := map[int]bool{}
	for len(lineIdxes) < p.taskSize {
		pickIdx := rand.Intn(len(lines))
		if lineIdxesMap[pickIdx] {
			continue
		}
		lineIdxes = append(lineIdxes, pickIdx)
		lineIdxesMap[pickIdx] = true
	}
	sort.Ints(lineIdxes)
	randomLines := make([]lib.OCRLine, 0, len(lineIdxes))
	for _, lineIdx := range lineIdxes {
		randomLines = append(randomLines, lines[lineIdx])
	}
	// Run in the background, the user does not have to wait for our
	// caching
	go lib.LineCache.CacheLines(randomLines, p.ident)
	p.writeMessage("lines", randomLines)
}

func (p *lineProducer) streamLines() {
	c, _ := p.resp.(http.CloseNotifier)
	closer := c.CloseNotify()
	for {
		select {
		case progMsg, ok := <-p.progChan:
			if !ok {
				p.progChan = nil
				break
			}
			p.writeMessage("progress", progMsg)
		case allLines, ok := <-p.lineChan:
			if !ok {
				p.lineChan = nil
				break
			}
			log.Printf(
				"Got lines for %s, picking %d at random and caching them.",
				p.ident, p.taskSize)
			p.handleLines(allLines)
		case <-closer:
			return
		}
		if p.progChan == nil && p.lineChan == nil {
			return
		}
	}
}
