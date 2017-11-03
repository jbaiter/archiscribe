package web

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

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
		taskChan <- task
		result := <-task.ResultChan
		js, _ := json.Marshal(result)
		w.Header().Add("Content-Type", "application/json")
		if result.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write(js)
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
			// Downloading is only 50% of the overall progress
			progMsg.Progress = 0.5 * progMsg.Progress
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
			log.Printf(
				"Got lines for %s, picking %d at random and caching them.",
				ident, taskSize)
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
				lib.LineCache.CacheLine(
					strings.Replace(allLines[pickIdx].ImageURL, ".jpg", ".png", -1),
					lib.MakeLineIdentifier(ident, allLines[pickIdx]))
				progMsg := lib.ProgressMessage{
					Step:     "cache",
					Progress: 0.50 + (0.50 * (float64(len(lineIdxes)) / float64(taskSize))),
				}
				progJSON, _ := json.Marshal(progMsg)
				fmt.Fprintf(resp, "event: progress\n")
				fmt.Fprintf(resp, "data: %s\n\n", progJSON)
				f.Flush()
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
func Serve(port int, repoPath string) {
	box := packr.NewBox("../client/dist")

	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write(box.Bytes("index.html"))
	})
	router.GET("/api/lines/:year", ProduceLines)
	router.POST("/api/transcriptions", SubmitTranscription)

	// NOTE: This is a bit clumsy, since Box.Open does not return an error
	// that is recognized by os.IsNotExit, which is why we have to pass
	// our own logic to return a 404 error for non-existing files.
	fileServer := http.FileServer(box)
	router.NotFound = func(w http.ResponseWriter, r *http.Request) {
		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
		}
		if box.Has(path.Clean(upath)) {
			fileServer.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}

	go lib.GitWatcher(repoPath, taskChan)

	fmt.Printf("Serving on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
