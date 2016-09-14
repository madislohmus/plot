package main

import (
	"encoding/hex"
	"image"
	"image/color"
	"math"
	"math/rand"
)

func extremes(array []int64) (int64, int64) {
	if len(array) == 0 {
		return 0, 0
	}
	min := int64(math.MaxInt64)
	max := int64(math.MinInt64)
	for _, element := range array {
		if element > max {
			max = element
		}
		if element < min {
			min = element
		}
	}
	return min, max
}

func getRandomColor() color.NRGBA {
	return color.NRGBA{
		R: uint8(rand.Intn(255)),
		G: uint8(rand.Intn(255)),
		B: uint8(rand.Intn(255)),
		A: 128}
}

func getColorFromHTML(colorString string) color.NRGBA {
	if len(colorString) != 6 {
		return getRandomColor()
	}
	rHex := colorString[0:2]
	gHex := colorString[2:4]
	bHex := colorString[4:6]
	r, _ := hex.DecodeString(rHex)
	g, _ := hex.DecodeString(gHex)
	b, _ := hex.DecodeString(bHex)
	return color.NRGBA{
		R: uint8(r[0]),
		G: uint8(g[0]),
		B: uint8(b[0]),
		A: 128}
}

func getSectorBoundingBox(startAngle, endAngle float64, size int64) (int64, int64, int64, int64) {
	var x1, x2, y1, y2 int64
	var xvalues, yvalues []int64
	compassAngles := []float64{0.0, math.Pi / 2, math.Pi, 3 * math.Pi / 2}
	r := float64(float64(size) / 2.0)
	xvalues = append(xvalues, int64(r*math.Cos(startAngle)))
	xvalues = append(xvalues, int64(r*math.Cos(endAngle)))
	xvalues = append(xvalues, int64(0))
	yvalues = append(yvalues, int64(-r*math.Sin(startAngle)))
	yvalues = append(yvalues, int64(-r*math.Sin(endAngle)))
	yvalues = append(yvalues, int64(0))
	x1, x2 = extremes(xvalues)
	y1, y2 = extremes(yvalues)
	if startAngle <= compassAngles[0] && endAngle >= compassAngles[0] {
		x2 = int64(r)
	}
	if startAngle <= compassAngles[1] && endAngle >= compassAngles[1] {
		y1 = int64(-r)
	}
	if startAngle <= compassAngles[2] && endAngle >= compassAngles[2] {
		x1 = int64(-r)
	}
	if startAngle <= compassAngles[3] && endAngle >= compassAngles[3] {
		y2 = int64(r)
	}
	return int64(float64(x1) + r), int64(float64(x2) + r), int64(float64(y1) + r), int64(float64(y2) + r)
}

func drawLine(i *image.NRGBA, c color.NRGBA, x1, x2, y1, y2 int64) {
	if math.Abs(float64(y2-y1)) > math.Abs(float64(x2-x1)) {
		if y2 < y1 {
			temp := y1
			y1 = y2
			y2 = temp
			temp = x1
			x1 = x2
			x2 = temp
		}
		for m := y1; m < y2; m++ {
			x := int64(float64(m-y1)/float64(y2-y1)*float64(x2-x1)) + x1
			i.Set(int(x), int(m), c)
		}
	} else {
		if x2 < x1 {
			temp := y1
			y1 = y2
			y2 = temp
			temp = x1
			x1 = x2
			x2 = temp
		}
		for m := x1; m < x2; m++ {
			y := int64(float64(m-x1)/float64(x2-x1)*float64(y2-y1)) + y1
			i.Set(int(m), int(y), c)
		}
	}
}
