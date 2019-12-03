package main

import (
	"fmt"
	"image/color"
	"image/jpeg"
	"io"
)

func countColors(r io.ReadCloser) (int, error) {
	img, err := jpeg.Decode(r)
	if err != nil {
		return 0, err
	}

	bounds := img.Bounds()
	colorCount := make(map[color.YCbCr]int64) //, (bounds.Max.X * bounds.Max.Y))
	for xi := 0; xi < bounds.Max.X; xi++ {
		for yi := 0; yi < bounds.Max.Y; yi++ {
			at := img.At(xi, yi)
			switch at.(type) {
			case color.YCbCr:
				ycbcr := at.(color.YCbCr)
				colorCount[ycbcr]++
			default:
				return 0, fmt.Errorf("image is not YCbCr")
			}
		}
	}
	return len(colorCount), nil
}
