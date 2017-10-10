package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"

	"github.com/bitly/go-simplejson"
	"gopkg.in/cheggaaa/pb.v2"
)

var pagePat = regexp.MustCompile(`<page width="(\d+)" height="(\d+)".+?>`)
var linePat = regexp.MustCompile(`<line .+?l="(\d+)" t="(\d+)" r="(\d+)" b="(\d+)">`)

// IDCache is a global cache for suitable identifiers
var IDCache *Cache

// OCRLine contains information about an OCR line
type OCRLine struct {
	ImageURL         string `json:"line"`
	PreviousImageURL string `json:"previous,omitempty"`
	NextImageURL     string `json:"next,omitempty"`
	Transcription    string `json:"transcription,omitempty"`
}

// TaskDefinition encodes a finished transcription along with author information
type TaskDefinition struct {
	Identifier string           `json:"id"`
	Lines      []OCRLine        `json:"lines"`
	Author     string           `json:"author,omitempty"`
	Comment    string           `json:"comment,omitempty"`
	Metadata   *simplejson.Json `json:"metadata"`
}

// ProgressReader wraps another reader and exposes progress information
type ProgressReader struct {
	proxiedReader io.Reader
	BytesRead     int64
}

// NewProgressReader creates a new ProgressReader from a given Reader
func NewProgressReader(proxied io.Reader) *ProgressReader {
	return &ProgressReader{proxied, 0}
}

func (r *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = r.proxiedReader.Read(p)
	if n > 0 {
		r.BytesRead += int64(n)
	} else {
		r.BytesRead += int64(len(p))
	}
	return n, err
}

// CacheLines caches three volumes for each year to disk
func CacheLines(cachePath string, year int, printProgress bool) string {
	yearPath := path.Join(cachePath, strconv.Itoa(year))
OuterCache:
	for {
		entry := IDCache.Random(year)
		ident := entry.Identifier
		isFrak, _ := IsFraktur(ident)
		if !isFrak {
			continue
		}
		progChan, lineChan, err := FetchLines(ident)
		if err != nil {
			log.Printf("Error while getting OCR for %s: %+v", ident, err)
			continue
		}
		log.Printf("Caching lines for %d from %s", year, ident)
		var progBar *pb.ProgressBar
		if printProgress {
			progBar = pb.New(100)
			progBar.SetWidth(80)
			progBar.Start()
		}
		for {
			select {
			case prog := <-progChan:
				if prog.Error != nil {
					log.Printf("Error while getting lines for %s: %+v", ident, prog.Error)
					continue OuterCache
				} else if printProgress {
					progBar.SetCurrent(int64(prog.Progress * 100))
				}
			case lines := <-lineChan:
				filePath := path.Join(yearPath, ident+".json")
				lineJSON, _ := json.Marshal(lines)
				ioutil.WriteFile(filePath, lineJSON, 0644)
				if printProgress {
					progBar.Finish()
				}
				return filePath
			}
		}
	}
}

// TODO: Holy shit, maybe that whole caching thing is completely unneccessary
//       and the slowness was just due to Python >_<
func cacheWatcher(basePath string) (map[int]chan string, error) {
	cacheChannels := map[int]chan string{}
	cacheFiles := map[int][]string{}
	//bufferedIds := map[string]bool{}
	yearDirs, err := ioutil.ReadDir(basePath)
	if err != nil {
		return nil, err
	}
	for _, yearDir := range yearDirs {
		if !yearDir.IsDir() {
			continue
		}
		year, _ := strconv.Atoi(yearDir.Name())
		yearPath := path.Join(basePath, yearDir.Name())
		dirContent, _ := ioutil.ReadDir(yearPath)
		cacheFiles[year] = make([]string, len(dirContent))
		for _, f := range dirContent {
			if f.IsDir() || path.Ext(f.Name()) != ".json" {
				continue
			}
			cacheFiles[year] = append(cacheFiles[year], path.Join(yearPath, f.Name()))
		}
		// Fill up cache
		for len(cacheFiles[year]) < 3 {
			cacheFiles[year] = append(cacheFiles[year], CacheLines(basePath, year, true))
		}
	}
	//go func() {
	// TODO: Create SelectCases with the remaining line
	// TODO: Select on the cases, when one is selected fetch another item
	//		 for that year, update the case and continue selecting
	//}()
	return cacheChannels, nil
	/*
		cacheMap := map[chan []OCRLine]int{}
		cases := make([]reflect.SelectCase, len(cacheMap))
		for year := range idCache {
			ch := make(chan []OCRLine)
			cacheMap[ch] = year
			cases = append(cases, reflect.SelectCase{
				Dir: reflect.SelectSend, Chan: reflect.ValueOf(ch),
				Send: reflect.New(nil)}) // TODO: Load from next cached lines file
		}
		for {
			// Wait for channels to become free
			chosen, recv, recvOk := reflect.Select(cases)
			year := cacheMap[cases[chosen].Chan.Close]

		}
	*/
}

func saveLine(lineURL string, baseDir string, transcription string) error {
	return nil
}

func createReadme(repoPath string) string {
	return ""
}

// InitCache initializes global identifier cache
func InitCache() {
	if _, err := os.Stat("./identifiers.json"); err != nil {
		fmt.Println("Caching identifiers...")
		cache, err := CacheIdentifiers("./identifiers.json")
		if err != nil {
			panic(err)
		}
		IDCache = cache
	} else {
		IDCache = LoadCache("./identifiers.json")
	}
}
