package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/alrs/ehloehmo"
	"log"
	"net/url"
	"os"
	"sync"

	bolt "github.com/etcd-io/bbolt"
)

const boltPath = "/tmp/topcolors.db"

var outPath = "/tmp/topcolors.csv"
var inPath = "testdata/input.txt"
var debris = false

const resultBucket = "urls"
const failBucket = "fail"

type resultPair struct {
	url      *url.URL
	topThree []byte
}

func init() {
	flag.StringVar(&outPath, "out", outPath, "path to csv output")
	flag.StringVar(&inPath, "in", inPath, "path to input data")
	flag.BoolVar(&debris, "debris", debris, "retain boltdb database on exit")
	flag.Parse()
}

func recordFailure(u *url.URL, db *bolt.DB) error {
	log.Printf("recording failure for %s", u.String())
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(failBucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(u.String()), []byte("T"))
	})
	return err
}

func recordResult(rp resultPair, db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(resultBucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(rp.url.String()), rp.topThree)
	})
}

func publishResults(output *os.File, db *bolt.DB) error {
	tx, err := db.Begin(false)
	if err != nil {
		return fmt.Errorf("Begin():%v", err)
	}
	defer tx.Rollback()
	fb := tx.Bucket([]byte(resultBucket))
	if fb == nil {
		return fmt.Errorf("bucket %s not found", resultBucket)
	}

	writer := csv.NewWriter(output)
	defer writer.Flush()

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(resultBucket))
		return b.ForEach(func(k, v []byte) error {
			colors := []string{}
			err := json.Unmarshal(v, &colors)
			if err != nil {
				return fmt.Errorf("Unmarshal():%v", err)
			}
			line := append([]string{string(k)}, colors...)
			err = writer.Write(line)
			if err != nil {
				return fmt.Errorf("csv Write():%v", err)
			}
			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("resultBucket ForEach(): %v", err)
	}
	log.Printf("result saved to file: %s", output.Name())
	return nil
}

func openFiles() (*os.File, *os.File, *bolt.DB, error) {
	input, err := os.Open(inPath)
	if err != nil {
		return nil, nil, nil, err
	}
	output, err := os.Create(outPath)
	if err != nil {
		return nil, nil, nil, err
	}
	db, err := bolt.Open(boltPath, 0644, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	return input, output, db, nil
}

func createBuckets(db *bolt.DB) error {
	tx, err := db.Begin(true)
	if err != nil {
		return fmt.Errorf("Begin():%v", err)
	}
	defer tx.Commit()
	_, err = tx.CreateBucket([]byte(resultBucket))
	if err != nil {
		return fmt.Errorf("resultBucket CreateBucket():%v", err)
	}
	_, err = tx.CreateBucket([]byte(failBucket))
	if err != nil {
		return fmt.Errorf("failBucket CreateBucket():%v", err)
	}
	return nil
}

func writeRoutine(resultCh chan resultPair, failCh chan *url.URL,
	doneCh chan struct{}, db *bolt.DB) {
loop:
	for {
		select {
		case failURL := <-failCh:
			err := recordFailure(failURL, db)
			if err != nil {
				log.Fatalf("recordFailure():%v", err)
			}
		case result := <-resultCh:
			err := recordResult(result, db)
			if err != nil {
				log.Fatalf("recordResult():%v", err)
			}
		case <-doneCh:
			close(failCh)
			close(resultCh)
			break loop
		}
	}
}

func readRoutine(input *os.File, resultCh chan resultPair,
	failCh chan *url.URL, db *bolt.DB) {
	wg := sync.WaitGroup{}
	scanner := bufio.NewScanner(input)
	var lineNum uint64
	pool := make(chan struct{}, 4)
	for scanner.Scan() {
		text := scanner.Text()
		pool <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-pool }()
			// read a URL from the input file
			lineNum++
			u, err := url.Parse(text)
			if err != nil {
				log.Printf("parse error line %d: %v", lineNum, err)
				return
			}

			// if the file extension isn't .jpeg, bail out
			if !ehloehmo.IsJPEG(u) {
				log.Printf("%s does not look to be a jpeg, ignoring", u.String())
				return
			}

			// already ingested?
			doneKey := []byte{}
			err = db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(resultBucket))
				if b == nil {
					log.Fatalf("missing bucket %q", resultBucket)
				}
				doneKey = b.Get([]byte(u.String()))
				return nil
			})
			if err != nil {
				log.Fatalf("View():%v", err)
			}

			if string(doneKey) != "" {
				log.Printf("%s already ingested, ignoring.", u.String())
				return
			}

			// already seen failure?
			failKey := []byte{}
			err = db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(failBucket))
				if b == nil {
					log.Fatalf("missing bucket %q", failBucket)
				}
				failKey = b.Get([]byte(u.String()))
				return nil
			})
			if err != nil {
				log.Fatalf("View():%v", err)
			}

			if string(failKey) != "" {
				log.Printf("%s is a known bad URL, ignoring.", u.String())
				return
			}

			jpg, err := ehloehmo.GetFile(u)
			if err != nil {
				log.Print(err)
				failCh <- u
				return
			}
			defer jpg.Close()

			cc, err := ehloehmo.ColorCounts(jpg)
			if err != nil {
				log.Printf("error counting colors in %s: %v", u.String(), err)
				failCh <- u
				return
			}

			sorted := ehloehmo.SortColorCounts(cc)
			topThree, err := sorted.CSVReady()
			if err != nil {
				log.Fatalf("CSVReady():%v", err)
			}

			ttJSON, err := json.Marshal(topThree)
			if err != nil {
				log.Fatalf("Marshal():%v", err)
			}

			resultCh <- resultPair{u, ttJSON}
		}()
	}

	wg.Wait()
}

func main() {
	input, output, db, err := openFiles()
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()
	defer output.Close()
	if !debris {
		defer os.RemoveAll(boltPath)
	}
	defer db.Close()

	err = createBuckets(db)
	if err != nil {
		log.Fatal(err)
	}

	failCh := make(chan *url.URL, 1)
	resultCh := make(chan resultPair, 1)
	doneCh := make(chan struct{})

	go writeRoutine(resultCh, failCh, doneCh, db)
	readRoutine(input, resultCh, failCh, db)

	// close down the writer goroutine
	close(doneCh)

	err = publishResults(output, db)
	if err != nil {
		log.Fatalf("publishResults(): %v", err)
	}
}
