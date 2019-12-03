package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/alrs/ehloehmo"
	"log"
	"net/url"
	"os"
	"sync"

	bolt "github.com/etcd-io/bbolt"
)

const outPath = "/tmp/topcolors.csv"
const boltPath = "/tmp/topcolors.db"
const inPath = "testdata/input.txt"

const colorBucket = "urls"
const failBucket = "fail"

type dbLock struct {
	sync.Mutex
}

func recordFailure(u *url.URL, b *bolt.Bucket) error {
	log.Printf("recording failure of %s", u.String())
	err := b.Put([]byte(u.String()), []byte("T"))
	if err != nil {
		return err
	}
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
	_, err = tx.CreateBucket([]byte(colorBucket))
	if err != nil {
		return fmt.Errorf("colorBucket CreateBucket():%v", err)
	}
	_, err = tx.CreateBucket([]byte(failBucket))
	if err != nil {
		return fmt.Errorf("failBucket CreateBucket():%v", err)
	}
	return nil
}

func main() {
	input, output, db, err := openFiles()
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()
	defer output.Close()
	defer os.Remove(boltPath)
	defer db.Close()

	err = createBuckets(db)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(input)
	var lineNum uint64
	for scanner.Scan() {
		text := scanner.Text()
		func() {

			// one transaction per loop
			tx, err := db.Begin(true)
			if err != nil {
				log.Fatalf("Begin():%v", err)
			}
			defer tx.Commit()

			// ready the failbucket
			fb := tx.Bucket([]byte(failBucket))
			if fb == nil {
				log.Fatal(fmt.Errorf("Bucket %s not found", failBucket))
			}

			// ready the colorbucket
			cb := tx.Bucket([]byte(colorBucket))
			if cb == nil {
				log.Fatal(fmt.Errorf("Bucket %s not found", colorBucket))
			}

			// read a URL from the input file
			lineNum++
			u, err := url.Parse(text)
			if err != nil {
				log.Printf("parse error line %d: %v", lineNum, err)
				return
			}

			// already ingested?
			doneKey := cb.Get([]byte(u.String()))
			if string(doneKey) != "" {
				log.Printf("%s already ingested, ignoring.", u.String())
				return
			}

			// already seen failure?
			failKey := fb.Get([]byte(u.String()))
			if string(failKey) != "" {
				log.Printf("%s is a known bad URL, ignoring.", u.String())
				return
			}

			// if the file extension isn't .jpeg, bail out
			if !ehloehmo.IsJPEG(u) {
				log.Printf("%s does not look to be a jpeg, ignoring", u.String())
				recordFailure(u, fb)
				return
			}

			jpg, err := ehloehmo.GetFile(u)
			if err != nil {
				log.Print(err)
				err := recordFailure(u, fb)
				if err != nil {
					log.Fatal(err)
				}
				return
			}
			defer jpg.Close()

			cc, err := ehloehmo.ColorCounts(jpg)
			if err != nil {
				log.Printf("error counting colors in %s: %v", u.String(), err)
				err := recordFailure(u, fb)
				if err != nil {
					log.Fatal(err)
				}
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

			err = cb.Put([]byte(u.String()), ttJSON)
			if err != nil {
				log.Fatalf("colorBucket Put(): %v", err)
			}
		}()
	}

	tx, err := db.Begin(false)
	if err != nil {
		log.Fatalf("Begin():%v", err)
	}
	defer tx.Rollback()
	fb := tx.Bucket([]byte(colorBucket))
	if fb == nil {
		log.Fatal(fmt.Errorf("Bucket %s not found", colorBucket))
	}

	writer := csv.NewWriter(output)
	defer writer.Flush()

	err = fb.ForEach(func(k, v []byte) error {
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
	if err != nil {
		log.Fatalf("colorBucket ForEach(): %v", err)
	}
}
