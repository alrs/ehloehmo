package ehloehmo

import (
	"fmt"
	"image/color"
	"image/jpeg"
	"io"
	"sort"
)

type ColorCount map[color.YCbCr]int

// A Pair is a sortable struct representation of the ColorCount map keys
// and values.
type Pair struct {
	Key   color.YCbCr
	Value int
}

// HexKey returns the RGB hex value of a Pair Key.
func (p *Pair) HexKey() string {
	r, g, b := color.YCbCrToRGB(p.Key.Y, p.Key.Cb, p.Key.Cr)
	return fmt.Sprintf("%02x%02x%02x", r, g, b)
}

// PairList is a slice of Pair structs.
type PairList []Pair

// CSVReady returns a string slice ready for CSV marshaling.
func (pl PairList) CSVReady() ([]string, error) {
	out := make([]string, 3)
	if len(pl) < 1 {
		return out,
			fmt.Errorf("a populated pairlist should have at least one color")
	}
	for i := 0; i < 3; i++ {
		out[i] = pl[len(pl)-(1+i)].HexKey()
	}
	return out, nil
}

// Len helps PairList satisfy the sort.Sort() interface.
func (pl PairList) Len() int { return len(pl) }

// Less helps PairList satisfy the sort.Sort() interface.
func (pl PairList) Less(i, j int) bool {
	return pl[i].Value < pl[j].Value
}

// Swap helps PairList satisfy the sort.Sort() interface.
func (pl PairList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

// SortColorCounts takes an unordered ColorCount map, converts it to a
// PairList, and returns a PairList sorted by key.
func SortColorCounts(cc ColorCount) PairList {
	pl := make(PairList, len(cc))
	i := 0
	for k, v := range cc {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(pl)
	return pl
}

// ColorCounts iterates over the pixels in a .JPEG and returns a map with
// color.YCbCr values as keys, and number of pixels counted per color key
// as values.
func ColorCounts(r io.ReadCloser) (ColorCount, error) {
	img, err := jpeg.Decode(r)
	if err != nil {
		return ColorCount{}, err
	}

	bounds := img.Bounds()
	cc := make(ColorCount)
	for xi := 0; xi < bounds.Max.X; xi++ {
		for yi := 0; yi < bounds.Max.Y; yi++ {
			at := img.At(xi, yi)
			// this should always come up as color.YCbCr with .jpeg images, but
			// it's a big Internet out there
			switch at.(type) {
			case color.YCbCr:
				ycbcr := at.(color.YCbCr)
				cc[ycbcr]++
			default:
				return cc, fmt.Errorf("image is not YCbCr")
			}
		}
	}
	return cc, nil
}
