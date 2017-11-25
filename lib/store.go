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

// TranscriptionStore offers an interface to the transcriptions
type TranscriptionStore struct {
	basePath string
	repo     *GitRepo
}

// Transcription holds all information about a transcription
type Transcription struct {
	Identifier string     `json:"id"`
	Title      string     `json:"title"`
	Year       int        `json:"year"`
	Manifest   string     `json:"manifest"`
	Lines      []OCRLine  `json:"lines,omitempty"`
	History    []LogEntry `json:"history,omitempty"`
	NumLines   int        `json:"numLines,omitempty"`
}

var lineNamePat = regexp.MustCompile(`(.+?)_([a-z0-9]{8})`)

// NewTranscriptionStore creates a new transcription store
func NewTranscriptionStore(path string) (*TranscriptionStore, error) {
	repo, err := GitOpen(path)
	if err != nil {
		return nil, err
	}
	return &TranscriptionStore{
		basePath: path,
		repo:     repo,
	}, nil
}

// Details retrieves a single Transcription by its identifier
func (s *TranscriptionStore) Details(ident string) *Transcription {
	var trans Transcription
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
	if err := json.Unmarshal(raw, &trans); err != nil {
		panic(err)
	}
	for idx, line := range trans.Lines {
		textPath := strings.Replace(metaPath, ".json", "_"+line.Identifier+".txt", -1)
		text, err := ioutil.ReadFile(textPath)
		if err != nil {
			panic(err)
		}
		trans.Lines[idx].Transcription = strings.TrimSpace(string(text))
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
	trans.History = log
	return &trans
}

// List all transcriptions
func (s *TranscriptionStore) List() []*Transcription {
	transPath := filepath.Join(s.basePath, "transcriptions")
	metaPaths, err := filepath.Glob(filepath.Join(transPath, "*", "*.json"))
	transcriptions := make([]*Transcription, 0, len(metaPaths))
	if err != nil {
		panic(err)
	}
	for _, metaPath := range metaPaths {
		trans := s.Details(strings.Replace(filepath.Base(metaPath), ".json", "", -1))
		trans.NumLines = len(trans.Lines)
		trans.Lines = trans.Lines[:0]
		if trans.Identifier != "" {
			transcriptions = append(transcriptions, trans)
		}
	}
	return transcriptions
}

func (s *TranscriptionStore) removeDeletedLines(trans Transcription) {
	basePath := filepath.Join(s.basePath, "transcriptions", strconv.Itoa(trans.Year))
	globPat := basePath + "/" + trans.Identifier + "*.png"
	lpaths, _ := filepath.Glob(globPat)
	for _, lpath := range lpaths {
		baseName := strings.TrimSuffix(filepath.Base(lpath), filepath.Ext(lpath))
		match := lineNamePat.FindStringSubmatch(baseName)
		if len(match) == 0 {
			continue
		}
		ident := match[2]
		found := false
		for _, line := range trans.Lines {
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

// Save a transcription
func (s *TranscriptionStore) Save(trans Transcription, author string, email string, comment string) (*Transcription, error) {
	if err := s.repo.CleanUp(); err != nil {
		return nil, err
	}
	if err := s.repo.Pull("origin", "master", true); err != nil {
		return nil, err
	}

	yearPath := filepath.Join(
		s.basePath, "transcriptions", strconv.Itoa(trans.Year))
	os.MkdirAll(yearPath, 0755)

	// Clear history, we don't persist it to disk
	trans.History = trans.History[:0]
	metaPath := filepath.Join(yearPath, trans.Identifier+".json")
	isUpdate := false
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		isUpdate = true
	}

	ident := trans.Identifier
	toRemove := make(map[string]bool)
	for idx, line := range trans.Lines {
		if line.Transcription == "" {
			// Not a transcribed line, removing from transcription
			toRemove[line.Identifier] = true
			continue
		}
		err := s.writeLineData(trans, line)
		if err != nil {
			return nil, err
		}
		// We don't store the transcriptions in the JSON
		trans.Lines[idx].Transcription = ""
	}
	filtered := make([]OCRLine, 0, len(trans.Lines)-len(toRemove))
	for _, line := range trans.Lines {
		if !toRemove[line.Identifier] {
			filtered = append(filtered, line)
		}
	}
	trans.Lines = filtered
	if err := LineCache.PurgeLines(ident); err != nil {
		return nil, err
	}
	if isUpdate {
		s.removeDeletedLines(trans)
	}

	// Write metadata
	metaOut, _ := os.Create(metaPath)
	enc := json.NewEncoder(metaOut)
	enc.SetIndent("", "  ")
	enc.Encode(trans)
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
		commitMessage = fmt.Sprintf("Corrected %s (%d)", trans.Identifier, trans.Year)
		changes, err := s.repo.Diff(true)
		if err != nil {
			return nil, err
		}
		if len(changes) == 0 {
			return s.Details(trans.Identifier), nil
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
			"Transcribed %d lines from %s (%d)", len(trans.Lines), trans.Identifier,
			trans.Year)
	}
	if comment != "" {
		commitMessage += ("\n" + comment)
	}
	if _, err := s.repo.Commit(commitMessage, author, email); err != nil {
		return nil, err
	}
	//s.repo.Push("origin", "master")
	return s.Details(trans.Identifier), nil
}

func (s *TranscriptionStore) writeLineData(trans Transcription, line OCRLine) error {
	basePath := filepath.Join(
		s.basePath, "transcriptions", strconv.Itoa(trans.Year),
		fmt.Sprintf("%s_%s", trans.Identifier, line.Identifier))
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

func (s *TranscriptionStore) createReadme() string {
	// FIXME: Counts are still broken
	transcriptions := s.List()
	sort.Slice(transcriptions, func(i, j int) bool {
		return transcriptions[i].Year < transcriptions[j].Year
	})

	numLinesTotal := 0
	yearCount := map[int]int{}
	decadeCount := map[int]int{}
	metaRows := [][]string{}
	for _, trans := range transcriptions {
		numLinesTotal += trans.NumLines
		decade := (trans.Year / 10) * 10
		yearCount[trans.Year] += trans.NumLines
		decadeCount[decade] += trans.NumLines
		archiveLink := fmt.Sprintf(
			"[%s](http://archive.org/details/%s)", trans.Identifier, trans.Identifier)
		manifestLink := fmt.Sprintf(
			"[Manifest](https://iiif.archivelab.org/iiif/%s/manifest.json)",
			trans.Identifier)
		miradorLink := fmt.Sprintf(
			"[Mirador](https://iiif.archivelab.org/iiif/%s)", trans.Identifier)
		metaRows = append(metaRows, []string{
			trans.Title, strconv.Itoa(trans.Year),
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
		"numWorks":    strconv.Itoa(len(transcriptions)),
		"numYears":    strconv.Itoa(len(years)),
		"decadeTable": decadesTable.String(),
		"yearTable":   yearsTable.String(),
		"worksTable":  metaTable.String(),
	})
	return out.String()
}
