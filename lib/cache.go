package lib

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Identifier Cache
// ==========================================================================

// IdentifierCacheEntry encodes cached information for a given Archive.org identifier
type IdentifierCacheEntry struct {
	Identifier string `json:"id"`
	NumPages   int    `json:"numPages"`
}

// IdentifierCache stores suitable identifiers
type IdentifierCache struct {
	path    string
	entries map[int][]IdentifierCacheEntry
}

// NewIdentifierCache constructs a new cache
func NewIdentifierCache(path string) *IdentifierCache {
	return &IdentifierCache{
		path:    path,
		entries: map[int][]IdentifierCacheEntry{}}
}

// LoadIdentifierCache loads a cache from a JSON file
func LoadIdentifierCache(path string) *IdentifierCache {
	cacheJSON, _ := ioutil.ReadFile(path)
	cache := IdentifierCache{path: path}
	json.Unmarshal(cacheJSON, &cache.entries)
	return &cache
}

// Write the cache to disk
func (c *IdentifierCache) Write() {
	cacheJSON, _ := json.Marshal(c.entries)
	ioutil.WriteFile(c.path, cacheJSON, 0644)
}

// Add a new entry to the cache
func (c *IdentifierCache) Add(ident string, numPages int, year int) {
	c.entries[year] = append(c.entries[year], IdentifierCacheEntry{
		Identifier: ident,
		NumPages:   numPages})
}

// Random returns a random identifier for a given year
func (c *IdentifierCache) Random(year int) IdentifierCacheEntry {
	pickIdx := rand.Intn(len(c.entries[year]))
	entry := c.entries[year][pickIdx]
	c.entries[year] = append(c.entries[year][:pickIdx], c.entries[year][pickIdx+1:]...)
	c.Write()
	return entry
}

// Line Image Cache
// ==========================================================================

// LineImageCache handles cached line images on disk
type LineImageCache struct {
	path string
}

// NewLineImageCache creates a new line image cache
func NewLineImageCache(cacheDir string) *LineImageCache {
	path := filepath.Join(cacheDir, "line_images")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
	cache := LineImageCache{
		path: filepath.Join(cacheDir, "line_images"),
	}
	go cache.purgeCacheWorker()
	return &cache
}

// Purges files older than 7 days, once a day
func (c *LineImageCache) purgeCacheWorker() {
	for {
		currentTime := time.Now()
		files, _ := ioutil.ReadDir(c.path)
		for _, finfo := range files {
			if currentTime.Sub(finfo.ModTime()).Hours() >= 7.0*24 {
				os.Remove(filepath.Join(c.path, finfo.Name()))
			}
		}
		time.Sleep(24 * time.Hour)
	}
}

// CacheLine downloads a line image and stores it on disk
func (c *LineImageCache) CacheLine(url string, id string) (string, error) {
	imgPath := filepath.Join(c.path, id+".png")
	imgOut, err := os.Create(imgPath)
	if err != nil {
		return "", err
	}
	defer imgOut.Close()
	imgResp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer imgResp.Body.Close()
	if _, err := io.Copy(imgOut, imgResp.Body); err != nil {
		return "", err
	}
	return imgPath, nil
}

// CacheLines caches all passed lines
func (c *LineImageCache) CacheLines(lines []OCRLine, ident string) {
	for _, line := range lines {
		c.CacheLine(line.ImageURL, MakeLineIdentifier(ident, line))
	}
}

// GetLinePath returns the file path for a given line image
func (c *LineImageCache) GetLinePath(id string) string {
	imgPath := filepath.Join(c.path, id+".png")
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		return ""
	}
	absPath, _ := filepath.Abs(imgPath)
	return absPath
}

// PurgeLines removes all cached line images that match the prefix
func (c *LineImageCache) PurgeLines(prefix string) error {
	lines, _ := filepath.Glob(filepath.Join(c.path, prefix+"*.png"))
	for _, fpath := range lines {
		if err := os.Remove(fpath); err != nil {
			return err
		}
	}
	return nil
}
