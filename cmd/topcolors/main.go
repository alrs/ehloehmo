package main

import (
	//	"fmt"
	"github.com/alrs/ehloehmo"
	//	"github.com/davecgh/go-spew/spew"
	//	"image/color"
	"log"
	"net/url"
	"os"
	"sync"
)

func main() {
	list, err := os.Open("testdata/input.txt")
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
			cc, err := ehloehmo.ColorCounts(b)
			if err != nil {
				log.Printf("error counting colors in %s: %v", u.String(), err)
				return
			}

			sorted := ehloehmo.SortColorCounts(cc)
			topThree, err := sorted.CSVReady()
			if err != nil {
				log.Printf("error sorting colors: %v", err)
				return
			}
			result := []string{u.String()}
			result = append(result, topThree...)
			log.Printf("%s", result)
		}()
	}
	wg.Wait()
}
