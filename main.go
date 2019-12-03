package main

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
)

const arbitraryMagicNumber = 999

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

func getJPEG(u *url.URL) (io.ReadCloser, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad http status on %s: %s", u, resp.Status)
	}
	return resp.Body, nil
}

func countColors(r io.ReadCloser) (int, error) {
	img, err := jpeg.Decode(r)
	if err != nil {
		return 0, err
	}

	bounds := img.Bounds()
	colorCount := make(map[color.YCbCr]int64) //, (bounds.Max.X * bounds.Max.Y))
	for xi := 0; xi < bounds.Max.X; xi++ {
		for yi := 0; yi < bounds.Max.Y; yi++ {
			ycbcr := img.At(xi, yi).(color.YCbCr)
			colorCount[ycbcr]++
		}
	}
	return len(colorCount), nil
}

func hexRGB(c color.YCbCr) {
	r, g, b := color.YCbCrToRGB(c.Y, c.Cb, c.Cr)
	fmt.Printf("r:%d g:%d b:%d hex: %02x%02x%02x\n", r, g, b, r, g, b)
}

func readURLS(r io.Reader, uc chan *url.URL) {
	defer close(uc)
	scanner := bufio.NewScanner(r)
	var lineNum uint64
	for scanner.Scan() {
		lineNum++
		u, err := url.Parse(scanner.Text())
		if err != nil {
			log.Print("parse error line %d: %v", lineNum, err)
			continue
		}
		uc <- u
	}
}

func main() {
	list, err := os.Open("input.txt")
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer list.Close()
	history := newHistory()
	urlChan := make(chan *url.URL, 10)

	go readURLS(list, urlChan)

	var wg sync.WaitGroup
	for {
		u := <-urlChan
		if u == nil {
			break
		}

		go func() {
			history.Lock()
			if history.check(u) {
				log.Printf("%s already processed, ignoring", u.String())
				history.Unlock()
				return
			}
			history.record(u)
			history.Unlock()

			wg.Add(1)
			defer wg.Done()
			b, err := getJPEG(u)
			if err != nil {
				log.Print(err)
				return
			}
			defer b.Close()
			count, err := countColors(b)
			if err != nil {
				log.Printf("error counting colors in %s: %v", u.String(), err)
				return
			}
			log.Printf("%s: %d colors", u.String(), count)
		}()
	}
	wg.Wait()
}
