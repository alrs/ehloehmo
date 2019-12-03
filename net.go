package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

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

func isJPEG(u *url.URL) bool {
	ep := u.EscapedPath()
	sp := strings.Split(ep, ".")
	ext := sp[len(sp)-1]
	if strings.ToLower(ext) == "jpg" || strings.ToLower(ext) == "jpeg" {
		return true
	}
	return false
}

func readURLS(r io.Reader, uc chan *url.URL) {
	defer close(uc)
	scanner := bufio.NewScanner(r)
	var lineNum uint64
	for scanner.Scan() {
		lineNum++
		u, err := url.Parse(scanner.Text())
		if err != nil {
			log.Printf("parse error line %d: %v", lineNum, err)
			continue
		}
		uc <- u
	}
}
