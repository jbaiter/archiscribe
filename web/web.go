package web

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/gobuffalo/packr"
	"github.com/julienschmidt/httprouter"

	"archiscribe/lib"
)

var taskChan = make(chan lib.TaskDefinition)

// SubmitTranscription handles user-submitted transcriptions
func SubmitTranscription(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var task lib.TaskDefinition
	err := json.NewDecoder(r.Body).Decode(&task)
	task.ResultChan = make(chan lib.SubmitResult)
	defer close(task.ResultChan)
	log.Printf("Received %d transcriptions for %s", len(task.Lines), task.Identifier)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%+v", err)))
	} else {
		log.Printf("Writing task to taskChan")
		taskChan <- task
		log.Printf("Waiting for commit")
		log.Printf("ResultChan has length %d", len(task.ResultChan))
		result := <-task.ResultChan
		log.Printf("Got result!")
		js, _ := json.Marshal(result)
		if result.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write(js)
		w.Header().Add("Content-Type", "application/json")
	}
}

// ProduceLines begins generating OCR lines for a given identifier
func ProduceLines(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	f, ok := resp.(http.Flusher)
	if !ok {
		http.Error(resp, "Streaming unsupported!",
			http.StatusInternalServerError)
		return
	}
	c, ok := resp.(http.CloseNotifier)
	if !ok {
		http.Error(resp, "close notification unsupported",
			http.StatusInternalServerError)
		return
	}
	year, _ := strconv.Atoi(ps.ByName("year"))
	var ident string
	var progChan chan lib.ProgressMessage
	var lineChan chan []lib.OCRLine
	for ident == "" {
		entry := lib.IDCache.Random(year)
		candidate := entry.Identifier
		isFrak, _ := lib.IsFraktur(candidate)
		if !isFrak {
			log.Printf("%s is not fraktur, continuing search\n", candidate)
			continue
		}
		var err error
		progChan, lineChan, err = lib.FetchLines(candidate)
		if err != nil {
			log.Printf("Error while getting OCR for %s: %+v", ident, err)
			continue
		}
		ident = candidate
	}
	log.Printf("Getting lines for %s", ident)
	headers := resp.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")
	closer := c.CloseNotify()

	metadata, _ := lib.GetMetadata(ident)
	metaJSON, _ := json.Marshal(metadata)
	fmt.Fprintf(resp, "event: metadata\n")
	fmt.Fprintf(resp, "data: %s\n\n", metaJSON)
	f.Flush()

	for {
		select {
		case progMsg, ok := <-progChan:
			if !ok {
				progChan = nil
				break
			}
			progJSON, _ := json.Marshal(progMsg)
			fmt.Fprintf(resp, "event: progress\n")
			fmt.Fprintf(resp, "data: %s\n\n", progJSON)
			f.Flush()
		case allLines, ok := <-lineChan:
			if !ok {
				lineChan = nil
				break
			}
			taskSize, _ := strconv.Atoi(req.URL.Query().Get("taskSize"))
			if taskSize == 0 {
				taskSize = 50
			}
			lineIdxes := make([]int, 0)
			lineIdxesMap := map[int]bool{}
			for len(lineIdxes) < taskSize {
				pickIdx := rand.Intn(len(allLines))
				if lineIdxesMap[pickIdx] {
					continue
				}
				lineIdxes = append(lineIdxes, pickIdx)
				lineIdxesMap[pickIdx] = true
			}
			sort.Ints(lineIdxes)
			randomLines := make([]lib.OCRLine, 0)
			for _, lineIdx := range lineIdxes {
				randomLines = append(randomLines, allLines[lineIdx])
			}
			lineJSON, _ := json.Marshal(randomLines)
			fmt.Fprintf(resp, "event: lines\n")
			fmt.Fprintf(resp, "data: %s\n\n", lineJSON)
			f.Flush()
		case <-closer:
			return
		}
		if progChan == nil && lineChan == nil {
			return
		}
	}
}

func addPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := prefix + r.URL.Path
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = p
		h.ServeHTTP(w, r2)
	})
}

// Serve the web application
func Serve(port int) {
	box := packr.NewBox("../client/dist")

	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write(box.Bytes("index.html"))
	})
	router.GET("/api/lines/:year", ProduceLines)
	router.POST("/api/transcriptions", SubmitTranscription)
	router.NotFound = http.FileServer(box).ServeHTTP

	go lib.GitWatcher("/tmp/temp-corpus", taskChan)

	fmt.Printf("Serving on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
