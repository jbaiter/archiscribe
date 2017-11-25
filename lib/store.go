package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/olekukonko/tablewriter"
)

// DocumentStore offers an interface to the transcriptions
type DocumentStore struct {
	basePath string
	repo     *GitRepo
}

// Document holds all information about a transcription document
type Document struct {
	Identifier string     `json:"id"`
	Title      string     `json:"title"`
	Year       int        `json:"year"`
	Manifest   string     `json:"manifest"`
	Lines      []OCRLine  `json:"lines,omitempty"`
	History    []LogEntry `json:"history,omitempty"`
	NumLines   int        `json:"numLines,omitempty"`
}

var lineNamePat = regexp.MustCompile(`(.+?)_([a-z0-9]{8})`)

// NewDocumentStore creates a new document store
func NewDocumentStore(path string) (*DocumentStore, error) {
	repo, err := GitOpen(path)
	if err != nil {
		return nil, err
	}
	return &DocumentStore{
		basePath: path,
		repo:     repo,
	}, nil
}

// Details retrieves a single Document by its identifier
func (s *DocumentStore) Details(ident string) *Document {
	var doc Document
	transPath := filepath.Join(s.basePath, "transcriptions")
	globPath := filepath.Join(transPath, "*", ident+".json")
	metaPaths, err := filepath.Glob(globPath)
	if err != nil {
		panic(err)
	}
	if len(metaPaths) == 0 {
		return nil
	}
	metaPath := metaPaths[0]
	raw, err := ioutil.ReadFile(metaPath)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		panic(err)
	}
	for idx, line := range doc.Lines {
		textPath := strings.Replace(metaPath, ".json", "_"+line.Identifier+".txt", -1)
		text, err := ioutil.ReadFile(textPath)
		if err != nil {
			panic(err)
		}
		doc.Lines[idx].Transcription = strings.TrimSpace(string(text))
	}
	transFiles, err := filepath.Glob(strings.Replace(metaPath, ".json", ".*", -1))
	if err != nil {
		panic(err)
	}
	transPaths := make([]string, 0, len(transFiles))
	for _, tf := range transFiles {
		tp, _ := filepath.Rel(s.basePath, tf)
		transPaths = append(transPaths, tp)
	}
	log, err := s.repo.Log(transPaths...)
	if err != nil {
		panic(err)
	}
	doc.History = log
	return &doc
}

// List all documents
func (s *DocumentStore) List() []*Document {
	transPath := filepath.Join(s.basePath, "transcriptions")
	metaPaths, err := filepath.Glob(filepath.Join(transPath, "*", "*.json"))
	documents := make([]*Document, 0, len(metaPaths))
	if err != nil {
		panic(err)
	}
	for _, metaPath := range metaPaths {
		doc := s.Details(strings.Replace(filepath.Base(metaPath), ".json", "", -1))
		doc.NumLines = len(doc.Lines)
		doc.Lines = doc.Lines[:0]
		if doc.Identifier != "" {
			documents = append(documents, doc)
		}
	}
	return documents
}

func (s *DocumentStore) removeDeletedLines(doc Document) {
	basePath := filepath.Join(s.basePath, "transcriptions", strconv.Itoa(doc.Year))
	globPat := basePath + "/" + doc.Identifier + "*.png"
	lpaths, _ := filepath.Glob(globPat)
	for _, lpath := range lpaths {
		baseName := strings.TrimSuffix(filepath.Base(lpath), filepath.Ext(lpath))
		match := lineNamePat.FindStringSubmatch(baseName)
		if len(match) == 0 {
			continue
		}
		ident := match[2]
		found := false
		for _, line := range doc.Lines {
			if line.Identifier == ident {
				found = true
				break
			}
		}
		if !found {
			if err := s.repo.Remove(lpath); err != nil {
				panic(err)
			}
			if err := s.repo.Remove(strings.Replace(lpath, ".png", ".txt", -1)); err != nil {
				panic(err)
			}
		}
	}
}

// Save a document
func (s *DocumentStore) Save(doc Document, author string, email string, comment string) (*Document, error) {
	if err := s.repo.CleanUp(); err != nil {
		return nil, err
	}
	if err := s.repo.Pull("origin", "master", true); err != nil {
		return nil, err
	}

	yearPath := filepath.Join(
		s.basePath, "transcriptions", strconv.Itoa(doc.Year))
	os.MkdirAll(yearPath, 0755)

	// Clear history, we don't persist it to disk
	doc.History = doc.History[:0]
	metaPath := filepath.Join(yearPath, doc.Identifier+".json")
	isUpdate := false
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		isUpdate = true
	}

	ident := doc.Identifier
	toRemove := make(map[string]bool)
	for idx, line := range doc.Lines {
		if line.Transcription == "" {
			// Not a transcribed line, removing from document
			toRemove[line.Identifier] = true
			continue
		}
		err := s.writeLineData(doc, line)
		if err != nil {
			return nil, err
		}
		// We don't store the transcriptions in the JSON
		doc.Lines[idx].Transcription = ""
	}
	filtered := make([]OCRLine, 0, len(doc.Lines)-len(toRemove))
	for _, line := range doc.Lines {
		if !toRemove[line.Identifier] {
			filtered = append(filtered, line)
		}
	}
	doc.Lines = filtered
	if err := LineCache.PurgeLines(ident); err != nil {
		return nil, err
	}
	if isUpdate {
		s.removeDeletedLines(doc)
	}

	// Write metadata
	metaOut, _ := os.Create(metaPath)
	enc := json.NewEncoder(metaOut)
	enc.SetIndent("", "  ")
	enc.Encode(doc)
	metaOut.Close()
	if err := s.repo.Add(metaPath); err != nil {
		return nil, err
	}

	readme := s.createReadme()
	readmePath := filepath.Join(s.basePath, "README.md")
	readmeOut, _ := os.Create(readmePath)
	readmeOut.WriteString(readme)
	readmeOut.Close()
	if err := s.repo.Add(readmePath); err != nil {
		return nil, err
	}
	var commitMessage string
	if isUpdate {
		commitMessage = fmt.Sprintf("Corrected %s (%d)", doc.Identifier, doc.Year)
		changes, err := s.repo.Diff(true)
		if err != nil {
			return nil, err
		}
		if len(changes) == 0 {
			return s.Details(doc.Identifier), nil
		}
		numModified := 0
		numDeleted := 0
		for fname, change := range changes {
			if !strings.HasSuffix(fname, ".txt") {
				continue
			}
			if change == StatusModified {
				numModified++
			} else if change == StatusDeleted {
				numDeleted++
			}
		}
		if numModified > 0 {
			commitMessage += fmt.Sprintf(", updated %d", numModified)
		}
		if numDeleted > 0 {
			commitMessage += fmt.Sprintf(", deleted %d", numDeleted)
		}
		if numModified > 0 || numDeleted > 0 {
			commitMessage += " lines"
		}
	} else {
		commitMessage = fmt.Sprintf(
			"Transcribed %d lines from %s (%d)", len(doc.Lines), doc.Identifier,
			doc.Year)
	}
	if comment != "" {
		commitMessage += ("\n" + comment)
	}
	if _, err := s.repo.Commit(commitMessage, author, email); err != nil {
		return nil, err
	}
	//s.repo.Push("origin", "master")
	return s.Details(doc.Identifier), nil
}

func (s *DocumentStore) writeLineData(doc Document, line OCRLine) error {
	basePath := filepath.Join(
		s.basePath, "transcriptions", strconv.Itoa(doc.Year),
		fmt.Sprintf("%s_%s", doc.Identifier, line.Identifier))
	imgPath := basePath + ".png"
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		// Obtain image file
		cachedPath := LineCache.GetLinePath(line.Identifier)
		if cachedPath == "" {
			path, err := LineCache.CacheLine(line.ImageURL, line.Identifier)
			if err != nil {
				return err
			}
			cachedPath = path
		}

		// Move line image from cache into repository
		in, err := os.Open(cachedPath)
		if err != nil {
			return err
		}
		out, err := os.Create(imgPath)
		if err != nil {
			return err
		}
		io.Copy(out, in)
		in.Close()
		out.Close()
		if err := os.Remove(cachedPath); err != nil {
			return err
		}
		if err := s.repo.Add(imgPath); err != nil {
			return err
		}
	}

	// Write transcription
	transPath := basePath + ".txt"
	transOut, err := os.Create(transPath)
	if err != nil {
		return err
	}
	if _, err = transOut.WriteString(line.Transcription + "\n"); err != nil {
		return err
	}
	transOut.Close()
	return s.repo.Add(transPath)
}

func (s *DocumentStore) createReadme() string {
	// FIXME: Counts are still broken
	documents := s.List()
	sort.Slice(documents, func(i, j int) bool {
		return documents[i].Year < documents[j].Year
	})

	numLinesTotal := 0
	yearCount := map[int]int{}
	decadeCount := map[int]int{}
	metaRows := [][]string{}
	for _, doc := range documents {
		numLinesTotal += doc.NumLines
		decade := (doc.Year / 10) * 10
		yearCount[doc.Year] += doc.NumLines
		decadeCount[decade] += doc.NumLines
		archiveLink := fmt.Sprintf(
			"[%s](http://archive.org/details/%s)", doc.Identifier, doc.Identifier)
		manifestLink := fmt.Sprintf(
			"[Manifest](https://iiif.archivelab.org/iiif/%s/manifest.json)",
			doc.Identifier)
		miradorLink := fmt.Sprintf(
			"[Mirador](https://iiif.archivelab.org/iiif/%s)", doc.Identifier)
		metaRows = append(metaRows, []string{
			doc.Title, strconv.Itoa(doc.Year),
			archiveLink, fmt.Sprintf("%s/%s", manifestLink, miradorLink)})
	}

	var yearsTable bytes.Buffer
	var years []int
	for k := range yearCount {
		years = append(years, k)
	}
	sort.Ints(years)
	t := tablewriter.NewWriter(&yearsTable)
	t.SetAutoFormatHeaders(false)
	t.SetHeader([]string{"Year", "# lines"})
	t.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	t.SetCenterSeparator("|")
	for _, year := range years {
		t.Append([]string{strconv.Itoa(year), strconv.Itoa(yearCount[year])})
	}
	t.Render()

	var decadesTable bytes.Buffer
	var decades []int
	for k := range decadeCount {
		decades = append(decades, k)
	}
	sort.Ints(decades)
	t = tablewriter.NewWriter(&decadesTable)
	t.SetAutoFormatHeaders(false)
	t.SetHeader([]string{"Decade", "# lines"})
	t.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	t.SetCenterSeparator("|")
	for _, decade := range decades {
		t.Append([]string{strconv.Itoa(decade), strconv.Itoa(decadeCount[decade])})
	}
	t.Render()

	var metaTable bytes.Buffer
	t = tablewriter.NewWriter(&metaTable)
	t.SetAutoFormatHeaders(false)
	t.SetAutoWrapText(false)
	t.SetHeader([]string{"Title", "Date", "Archive.org", "IIIF"})
	t.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	t.SetCenterSeparator("|")
	t.AppendBulk(metaRows)
	t.Render()

	var out bytes.Buffer
	tmpl := template.Must(template.New("README.md").Parse(readmeTemplate))
	tmpl.Execute(&out, map[string]string{
		"numLines":    strconv.Itoa(numLinesTotal),
		"numWorks":    strconv.Itoa(len(documents)),
		"numYears":    strconv.Itoa(len(years)),
		"decadeTable": decadesTable.String(),
		"yearTable":   yearsTable.String(),
		"worksTable":  metaTable.String(),
	})
	return out.String()
}
