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

// fractional part of x
func fpart(x float64) float64 {
	if x < 0 {
		return 1 - (x - float64(int(x)))
	}
	return x - float64(int(x))
}

func rfpart(x float64) float64 {
	return 1 - fpart(x)
}

func drawLine(i *image.NRGBA, c color.NRGBA, x0, x1, y0, y1 float64) {
	alpha := c.A
	steep := math.Abs(float64(y1)-float64(y0)) > math.Abs(float64(x1)-float64(x0))
	if steep {
		temp := x0
		x0 = y0
		y0 = temp
		temp = x1
		x1 = y1
		y1 = temp
	}
	if x0 > x1 {
		temp := x0
		x0 = x1
		x1 = temp
		temp = y0
		y0 = y1
		y1 = temp
	}

	dx := x1 - x0
	dy := y1 - y0
	gradient := dy / dx

	// handle first endpoint
	xend := int(x0 + 0.5)
	yend := y0 + gradient*(float64(xend)-x0)
	xgap := rfpart(x0 + 1.5)
	xpxl1 := xend // this will be used in the main loop
	ypxl1 := int(yend)
	if steep {
		c.A = uint8(float64(alpha) * rfpart(yend) * xgap)
		i.Set(ypxl1, xpxl1, c)
		c.A = uint8(float64(alpha) * fpart(yend) * xgap)
		i.Set(ypxl1+1, xpxl1, c)
	} else {
		c.A = uint8(float64(alpha) * rfpart(yend) * xgap)
		i.Set(xpxl1, ypxl1, c)
		c.A = uint8(float64(alpha) * fpart(yend) * xgap)
		i.Set(xpxl1, ypxl1+1, c)
	}
	intery := yend + gradient // first y-intersection for the main loop

	// handle second endpoint
	xend = int(x1 + 0.5)
	yend = y1 + gradient*(float64(xend)-x1)
	xgap = fpart(x1 + 0.5)
	xpxl2 := xend //this will be used in the main loop
	ypxl2 := int(yend)
	if steep {
		c.A = uint8(float64(alpha) * rfpart(yend) * xgap)
		i.Set(ypxl2, xpxl2, c)
		c.A = uint8(float64(alpha) * fpart(yend) * xgap)
		i.Set(ypxl2+1, xpxl2, c)
	} else {
		c.A = uint8(float64(alpha) * rfpart(yend) * xgap)
		i.Set(xpxl2, ypxl2, c)
		c.A = uint8(float64(alpha) * fpart(yend) * xgap)
		i.Set(xpxl2, ypxl2+1, c)
	}

	// main loop
	if steep {
		for x := xpxl1 + 1; x <= xpxl2; x++ {
			c.A = uint8(float64(alpha) * rfpart(intery))
			i.Set(int(intery), x, c)
			c.A = uint8(float64(alpha) * fpart(intery))
			i.Set(int(intery)+1, x, c)
			intery = intery + gradient
		}
	} else {
		for x := xpxl1 + 1; x <= xpxl2; x++ {
			c.A = uint8(float64(alpha) * rfpart(intery))
			i.Set(x, int(intery), c)
			c.A = uint8(float64(alpha) * fpart(intery))
			i.Set(x, int(intery)+1, c)
			intery = intery + gradient
		}
	}
}
