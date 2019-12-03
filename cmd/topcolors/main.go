package main

import (
	"fmt"
	"github.com/alrs/ehloehmo"
	"image/color"
	"log"
	"net/url"
	"os"
	"sync"
)

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
	history := ehloehmo.NewHistory()
	urlChan := make(chan *url.URL, 10)

	go ehloehmo.ReadURLS(list, urlChan)

	var wg sync.WaitGroup
	for {
		u := <-urlChan
		if u == nil {
			break
		}

		go func() {
			if !ehloehmo.IsJPEG(u) {
				log.Printf("%s does not look to be a jpeg, ignoring", u.String())
				return
			}
			history.Lock()
			if history.Check(u) {
				log.Printf("%s already processed, ignoring", u.String())
				history.Unlock()
				return
			}
			history.Record(u)
			history.Unlock()

			wg.Add(1)
			defer wg.Done()
			b, err := ehloehmo.GetFile(u)
			if err != nil {
				log.Print(err)
				return
			}
			defer b.Close()
			count, err := ehloehmo.CountColors(b)
			if err != nil {
				log.Printf("error counting colors in %s: %v", u.String(), err)
				return
			}
			log.Printf("%s: %d colors", u.String(), count)
		}()
	}
	wg.Wait()
}
