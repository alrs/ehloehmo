package main

import (
	"fmt"
	"image/color"
	"image/jpeg"
	"log"
	"os"
)

func main() {
	f, err := os.Open("test.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()
	colorCount := make(map[color.YCbCr]int64, (bounds.Max.X * bounds.Max.Y))
	for xi := 0; xi < bounds.Max.X; xi++ {
		for yi := 0; yi < bounds.Max.Y; yi++ {
			ycbcr := img.At(xi, yi).(color.YCbCr)
			colorCount[ycbcr]++
		}
	}
	log.Print(len(colorCount))
}

func hexRGB(c color.YCbCr) {
	r, g, b := color.YCbCrToRGB(c.Y, c.Cb, c.Cr)
	fmt.Printf("r:%d g:%d b:%d hex: %02x%02x%02x\n", r, g, b, r, g, b)
}
