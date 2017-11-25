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
var store *lib.DocumentStore

// APIError is for errors that are returned via the API
type APIError struct {
	Err  error `json:"error"`
	Code int   `json:"code"`
}

func writeAPIError(err error, code int, w http.ResponseWriter) {
	apiErr := APIError{
		Err:  err,
		Code: code}
	out, _ := json.MarshalIndent(apiErr, "", "  ")
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	w.Write(out)
}

// SubmitDocument handles user-submitted documents
func SubmitDocument(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var task lib.TaskDefinition
	err := json.NewDecoder(r.Body).Decode(&task)
	task.ResultChan = make(chan lib.SubmitResult)
	defer close(task.ResultChan)
	if r.Method == "POST" {
		log.Printf(
			"Received %d transcriptions for %s",
			len(task.Document.Lines), task.Document.Identifier)
	} else {
		log.Printf("Received update for %s", task.Document.Identifier)
	}
	if err != nil {
		writeAPIError(err, 500, w)
	} else {
		stored, err := store.Save(task.Document, task.Author, task.Email, task.Comment)
		if err != nil {
			writeAPIError(err, 500, w)
			return
		}
		js, _ := json.MarshalIndent(stored, "", "  ")
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
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

// ListDocuments returns a list of all documents
func ListDocuments(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	documents := store.List()
	raw, err := json.Marshal(documents)
	if err != nil {
		log.Printf("%+v\n", err)
		resp.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.Header().Add("Content-Type", "application/json")
		resp.Write(raw)
	}
}

// GetDocument returns a single document
func GetDocument(resp http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	doc := store.Details(ps.ByName("ident"))
	log.Printf("Getting %s", ps.ByName("ident"))
	raw, err := json.Marshal(doc)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
	} else if doc.Identifier == "" {
		resp.WriteHeader(http.StatusNotFound)
	} else {
		resp.Header().Add("Content-Type", "application/json")
		resp.Write(raw)
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
	s, err := lib.NewDocumentStore(repoPath)
	if err != nil {
		panic(err)
	}
	store = s
	box := packr.NewBox("../client/dist")

	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write(box.Bytes("index.html"))
	})
	router.GET("/api/lines/:year", ProduceLines)
	router.GET("/api/documents", ListDocuments)
	router.POST("/api/documents", SubmitDocument)
	router.GET("/api/documents/:ident", GetDocument)
	router.PUT("/api/documents/:ident", SubmitDocument)

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
	fmt.Printf("Serving on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
