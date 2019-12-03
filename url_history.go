package ehloehmo

import (
	"hash/crc32"
	"net/url"
	"sync"
)

const arbitraryMagicNumber = 999

// URLHistory stores checksums of already-processed URLs.
type URLHistory struct {
	table *crc32.Table
	// minimize RAM usage for storing previously-processed URLs
	checksums map[uint32]struct{}
	sync.Mutex
}

// Check the URLHistory to see if a URL has already been visited.
func (uh *URLHistory) Check(u *url.URL) bool {
	c := crc32.Checksum([]byte(u.String()), uh.table)
	_, ok := uh.checksums[c]
	return ok
}

// Record a URL to the URLHistory
func (uh *URLHistory) Record(u *url.URL) {
	c := crc32.Checksum([]byte(u.String()), uh.table)
	uh.checksums[c] = struct{}{}
}

// NewHistory generates a ready-to-use URLHistory.
func NewHistory() *URLHistory {
	uh := URLHistory{}
	uh.checksums = make(map[uint32]struct{})
	uh.table = crc32.MakeTable(arbitraryMagicNumber)
	return &uh
}
