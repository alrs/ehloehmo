package ehloehmo

import (
	"os"
	"reflect"
	"testing"
)

func TestCSVReady(t *testing.T) {
	f, err := os.Open("testdata/four.jpg")
	if err != nil {
		t.Fatal(err)
	}
	cc, err := ColorCounts(f)
	if err != nil {
		t.Fatal(err)
	}
	scc := SortColorCounts(cc)
	got, err := scc.CSVReady()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"ffffff",
		"9d1644",
		"27a015",
	}
	tmpl := "expected:%q got:%q"
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf(tmpl, expected, got)
	}
	t.Logf(tmpl, expected, got)
}
