package main

import (
	"fmt"
	"image/color"
	"log"
	"net/url"
	"os"
	"sync"
)

const arbitraryMagicNumber = 999

func hexRGB(c color.YCbCr) {
	r, g, b := color.YCbCrToRGB(c.Y, c.Cb, c.Cr)
	fmt.Printf("r:%d g:%d b:%d hex: %02x%02x%02x\n", r, g, b, r, g, b)
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
			if !isJPEG(u) {
				log.Printf("%s does not look to be a jpeg, ignoring", u.String())
				return
			}
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
