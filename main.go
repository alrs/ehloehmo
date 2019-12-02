package main

import (
	"fmt"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"os"
)

func main() {
	f, err := os.Open("test.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	count, err := countColors(f)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(count)
}

func countColors(r io.Reader) (int, error) {
	img, err := jpeg.Decode(r)
	if err != nil {
		return 0, err
	}

	bounds := img.Bounds()
	colorCount := make(map[color.YCbCr]int64, (bounds.Max.X * bounds.Max.Y))
	for xi := 0; xi < bounds.Max.X; xi++ {
		for yi := 0; yi < bounds.Max.Y; yi++ {
			ycbcr := img.At(xi, yi).(color.YCbCr)
			colorCount[ycbcr]++
		}
	}
	return len(colorCount), nil
}

func hexRGB(c color.YCbCr) {
	r, g, b := color.YCbCrToRGB(c.Y, c.Cb, c.Cr)
	fmt.Printf("r:%d g:%d b:%d hex: %02x%02x%02x\n", r, g, b, r, g, b)
}
