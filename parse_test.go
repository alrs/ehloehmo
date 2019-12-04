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
	expected := []string{"#ffffff",
		"#9d1644",
		"#27a015",
	}
	tmpl := "expected:%q got:%q"
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf(tmpl, expected, got)
	}
	t.Logf(tmpl, expected, got)
}

func TestColorCounts(t *testing.T) {
	cases := map[string]bool{
		"testdata/bad.jpg":  false,
		"testdata/good.jpg": true,
		"testdata/four.jpg": true,
	}
	for fn, expectedSuccess := range cases {
		f, err := os.Open(fn)
		if err != nil {
			t.Fatal(err)
		}
		_, err = ColorCounts(f)
		if expectedSuccess {
			if err != nil {
				t.Fatalf("%s expected success, got error: %v", fn, err)
			}
			t.Logf("%s got expected successful ColorCount()", fn)
		} else {
			if err == nil {
				t.Fatalf("%s expected failure, none seen.", fn)
			}
			t.Logf("%s expected failure, got failure.", fn)
		}
	}
}
