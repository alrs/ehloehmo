package main

import (
	"hash/crc32"
	"net/url"
	"sync"
)

type urlHistory struct {
	table     *crc32.Table
	checksums map[uint32]struct{}
	sync.Mutex
}

func (uh *urlHistory) check(u *url.URL) bool {
	c := crc32.Checksum([]byte(u.String()), uh.table)
	_, ok := uh.checksums[c]
	return ok
}

func (uh *urlHistory) record(u *url.URL) {
	c := crc32.Checksum([]byte(u.String()), uh.table)
	uh.checksums[c] = struct{}{}
}

func newHistory() *urlHistory {
	uh := urlHistory{}
	uh.checksums = make(map[uint32]struct{})
	uh.table = crc32.MakeTable(arbitraryMagicNumber)
	return &uh
}
