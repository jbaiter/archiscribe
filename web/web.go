package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
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
	year, _ := strconv.Atoi(ps.ByName("year"))
	taskSize, _ := strconv.Atoi(req.URL.Query().Get("taskSize"))
	lineProd, err := newLineProducer(resp, taskSize, year)
	if err != nil {
		log.Printf("%+v\n", err)
		resp.WriteHeader(http.StatusInternalServerError)
	} else {
		lineProd.produceLines()
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
