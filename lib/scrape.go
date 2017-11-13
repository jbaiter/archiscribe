package lib

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
	pb "gopkg.in/cheggaaa/pb.v2"
)

// Result stores a response from the Archive.org Scraping API
type Result struct {
	items  *simplejson.Json
	cursor string
	count  int
	total  int
}

// ProgressMessage contains progress information for the ABBYY parsing task
type ProgressMessage struct {
	Step       string  `json:"step"`
	Progress   float64 `json:"progress"`
	BytesTotal int64   `json:"bytesTotal,omitempty"`
	BytesRead  int64   `json:"bytesRead,omitempty"`
	PageNumber int     `json:"pageNumber,omitempty"`
	LineNumber int     `json:"lineNumber,omitempty"`
	Error      error   `json:"error,omitempty"`
}

func grabNext(totalOnly bool, count int, cursor string) (*Result, error) {
	params := url.Values{}
	params.Set("q", "mediatype:(texts) AND language:(German) AND "+
		"date:[1800-01-01 TO 1941-01-01]")
	params.Set("fields", "identifier,imagecount,year")
	if totalOnly {
		params.Set("total_only", "true")
	} else if cursor != "" {
		params.Set("cursor", cursor)
	}
	searchURL := "https://archive.org/services/search/v1/scrape?" + params.Encode()
	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, err
	} else if resp.StatusCode > 200 {
		return nil, fmt.Errorf("Status %d while getting %s", resp.StatusCode, searchURL)
	}
	defer resp.Body.Close()
	json, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Result{
		items:  json.Get("items"),
		cursor: json.Get("cursor").MustString(),
		count:  json.Get("count").MustInt(),
		total:  json.Get("total").MustInt()}, nil
}

func getYear(doc *simplejson.Json) int {
	val := doc.Get("year")
	if yearString, err := val.String(); err == nil {
		if year, err := strconv.Atoi(yearString); err == nil {
			return year
		}
	}
	if years, err := val.StringArray(); err == nil {
		for _, yearString := range years {
			if year, err := strconv.Atoi(yearString); err == nil {
				return year
			}
		}
	}
	return -1
}

// CacheIdentifiers scrapes the Archive.org API and caches information about
// relevant identifiers and their number of pages
func CacheIdentifiers(path string) (*IdentifierCache, error) {
	cache := NewIdentifierCache(path)
	res, err := grabNext(true, -1, "")
	if err != nil {
		return nil, err
	}
	numTotal := res.total

	progressBar := pb.New(numTotal)
	progressBar.SetWidth(80)
	progressBar.Start()
	processedCount := 0
	var cursor string
	for processedCount < numTotal {
		res, err := grabNext(false, 10000, cursor)
		if err != nil {
			return nil, err
		}
		for i := 0; i < res.count; i++ {
			itm := res.items.GetIndex(i)
			year := getYear(itm)
			numPages, err := itm.Get("imagecount").Int()
			if err != nil || numPages < 50 {
				continue
			}
			cache.Add(itm.Get("identifier").MustString(), numPages, year)
		}
		cursor = res.cursor
		processedCount += res.count
		progressBar.Add(res.count)
	}
	cache.Write()
	progressBar.Finish()
	return cache, nil
}

// GetMetadata fetches metadata for identifier from Archive.org
func GetMetadata(ident string) (*simplejson.Json, error) {
	metaURL := "https://archive.org/metadata/" + ident
	resp, err := http.Get(metaURL)
	if err != nil {
		return nil, err
	} else if resp.StatusCode > 200 {
		return nil, fmt.Errorf("Status %d while getting %s", resp.StatusCode, metaURL)
	}
	json, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return json.Get("metadata"), nil
}

// IsFraktur uses heuristics to determine wheter a given identifier is
// set in a Fraktur typeface
func IsFraktur(ident string) (bool, error) {
	ocrURL := fmt.Sprintf("https://archive.org/download/%s/%s_djvu.txt",
		ident, ident)
	resp, err := http.Get(ocrURL)
	if err != nil {
		return false, err
	} else if resp.StatusCode > 200 {
		return false, fmt.Errorf("Status %d while getting %s", resp.StatusCode, ocrURL)
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanWords)
	numIft := 0
	var curToken string
	for scanner.Scan() && numIft <= 5 {
		curToken = scanner.Text()
		if curToken == "ift" {
			numIft++
		}
	}
	return numIft > 5, nil
}

// GetStartPageNumber determines whether an identifier's first page has
// index 0 or 1
func GetStartPageNumber(ident string) int {
	infoURL := fmt.Sprintf("https://iiif.archivelab.org/iiif/%s$0/info.json",
		ident)
	resp, err := http.Get(infoURL)
	if err != nil {
		return 0
	} else if resp.StatusCode > 200 {
		return 1
	}
	return 0
}

func fetchLinesWorker(ident string, minLineWidth int, progressChan chan ProgressMessage, linesChan chan []OCRLine) {
	log.Printf("Getting ABBY OCR for %s\n", ident)
	boxURL := fmt.Sprintf("https://archive.org/download/%s/%s_abbyy.gz",
		ident, ident)
	resp, err := http.Get(boxURL)
	if err != nil {
		progressChan <- ProgressMessage{Error: err, Step: "fetch"}
		return
	} else if resp.StatusCode > 200 {
		progressChan <- ProgressMessage{
			Error: fmt.Errorf("Status %d while getting %s", resp.StatusCode, boxURL),
			Step:  "fetch"}
		return
	}
	numBytesTotal := resp.ContentLength
	log.Printf("Parsing lines from %s ABBYY OCR, has %d bytes\n", ident, numBytesTotal)
	progReader := NewProgressReader(resp.Body)
	gzReader, _ := gzip.NewReader(progReader)
	defer resp.Body.Close()
	defer gzReader.Close()
	lineScanner := bufio.NewScanner(gzReader)
	lineScanner.Split(bufio.ScanLines)
	buf := make([]byte, 64*1024)
	lineScanner.Buffer(buf, 16*1024*1024)
	lines := make([]OCRLine, 0)
	numLines := 0
	currentPageNo := GetStartPageNumber(ident) - 1
	pageWidth := -1
	pageHeight := -1
	progPercent := 0
	for lineScanner.Scan() {
		numLines++
		line := lineScanner.Text()
		if strings.Contains(line, "<page") {
			match := pagePat.FindStringSubmatch(line)
			pageWidth, _ = strconv.Atoi(match[1])
			pageHeight, _ = strconv.Atoi(match[2])
			currentPageNo++
		}
		if !strings.Contains(line, "<line") {
			continue
		}
		prct := int(100. * float64(progReader.BytesRead) / float64(numBytesTotal))
		if prct > progPercent {
			progPercent = prct
			progressChan <- ProgressMessage{
				Step:       "fetch",
				Progress:   float64(progReader.BytesRead) / float64(numBytesTotal),
				BytesTotal: numBytesTotal,
				BytesRead:  progReader.BytesRead,
				PageNumber: currentPageNo,
				LineNumber: numLines,
				Error:      nil,
			}
		}
		if currentPageNo <= 10 { // TODO: Should be dynamic or from constant
			continue
		}
		matches := linePat.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			x, _ := strconv.Atoi(match[1])
			y, _ := strconv.Atoi(match[2])
			lrx, _ := strconv.Atoi(match[3])
			lry, _ := strconv.Atoi(match[4])
			width := lrx - x
			height := lry - y
			relX := float64(x) / float64(pageWidth)
			relY := float64(y) / float64(pageHeight)
			if width < minLineWidth || (relX > 0.65 && relY > 0.90) {
				continue
			}
			iiifURL := fmt.Sprintf(
				"https://iiif.archivelab.org/iiif/%s$%d/%d,%d,%d,%d/full/0/default.png",
				ident, currentPageNo, x, y, width, height)
			if len(lines) > 0 {
				lines[len(lines)-1].NextImageURL = iiifURL
			}
			l := OCRLine{
				ImageURL: iiifURL,
			}
			if len(lines) > 0 {
				l.PreviousImageURL = lines[len(lines)-1].ImageURL
			}
			lines = append(lines, l)
		}
	}
	linesChan <- lines
	close(linesChan)
	close(progressChan)
}

// FetchLines fetches OCR lines for a given Archive.org identifier
func FetchLines(ident string) (chan ProgressMessage, chan []OCRLine) {
	progressChan := make(chan ProgressMessage)
	lineChan := make(chan []OCRLine)
	go fetchLinesWorker(ident, 200, progressChan, lineChan)
	return progressChan, lineChan
}
