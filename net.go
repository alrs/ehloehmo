package ehloehmo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GetFile retrieves a file from a URL.
func GetFile(u *url.URL) (io.ReadCloser, error) {
	timeout := time.Duration(30 * time.Second)
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad http status on %s: %s", u, resp.Status)
	}
	return resp.Body, nil
}

// IsJPEG checks the filename extension of a given URL to determine if it
// is a jpeg.
func IsJPEG(u *url.URL) bool {
	ep := u.EscapedPath()
	sp := strings.Split(ep, ".")
	ext := sp[len(sp)-1]
	if strings.ToLower(ext) == "jpg" || strings.ToLower(ext) == "jpeg" {
		return true
	}
	return false
}
