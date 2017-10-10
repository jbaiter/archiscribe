package lib

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
)

// CacheEntry encodes cached information for a given Archive.org identifier
type CacheEntry struct {
	Identifier string `json:"id"`
	NumPages   int    `json:"numPages"`
}

// Cache stores suitable identifiers
type Cache struct {
	path    string
	entries map[int][]CacheEntry
}

// NewCache constructs a new cache
func NewCache(path string) *Cache {
	return &Cache{
		path:    path,
		entries: map[int][]CacheEntry{}}
}

// LoadCache loads a cache from a JSON file
func LoadCache(path string) *Cache {
	cacheJSON, _ := ioutil.ReadFile(path)
	cache := Cache{path: path}
	json.Unmarshal(cacheJSON, &cache.entries)
	return &cache
}

// Write the cache to disk
func (c *Cache) Write() {
	cacheJSON, _ := json.Marshal(c.entries)
	ioutil.WriteFile(c.path, cacheJSON, 0644)
}

// Add a new entry to the cache
func (c *Cache) Add(ident string, numPages int, year int) {
	c.entries[year] = append(c.entries[year], CacheEntry{
		Identifier: ident,
		NumPages:   numPages})
}

// Random returns a random identifier for a given year
func (c *Cache) Random(year int) CacheEntry {
	pickIdx := rand.Intn(len(c.entries[year]))
	entry := c.entries[year][pickIdx]
	c.entries[year] = append(c.entries[year][:pickIdx], c.entries[year][pickIdx+1:]...)
	c.Write()
	return entry
}
