package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

func img(size int) image.Image {
	i := image.NewNRGBA(image.Rect(0, 0, size, size))
	for m := int(0.25 * float64(size)); m < int(0.75*float64(size)); m++ {
		for n := int(0.25 * float64(size)); n < int(0.75*float64(size)); n++ {
			i.Set(m, n, color.Black)
		}
	}
	return i
}

func getPieChartBuffer(size int, values []int64, colorsArray []color.NRGBA, donut bool) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	err := png.Encode(buffer, pieChart(size, values, colorsArray, donut))
	return buffer, err
}

func getBarChartBuffer(size int, values []int64, colorsArray []color.NRGBA) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	err := png.Encode(buffer, barChart(size, values, colorsArray))
	return buffer, err
}

func getLineChartBuffer(size int, values []int64, c color.NRGBA) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	err := png.Encode(buffer, lineChart(size, values, c))
	return buffer, err
}

func barChart(size int, values []int64, colorsArray []color.NRGBA) image.Image {
	min, max := extremes(values)
	i := image.NewNRGBA(image.Rect(0, 0, size, size))
	span := max
	if max < 0 {
		span = -max
	} else if min < 0 {
		span = max - min
	}
	zero := int64(0)
	if max < 0 && min < 0 {
		zero = int64(size)
	}
	if min < 0 {
		zero = int64(float64(size) * float64(-min) / float64(span))
	}
	c := getRandomColor()
	for nr, value := range values {
		thickness := float64(size) / float64(len(values))
		hStart := int64(float64(nr) * thickness)
		length := int64(float64(size) * float64(value) / float64(span))
		vStart := int64(size) - zero

		if len(colorsArray) > 0 {
			c = colorsArray[nr%len(colorsArray)]
		}
		for m := hStart; m <= int64(float64(hStart)+thickness); m++ {
			if length > 0 {
				for n := vStart; n > vStart-length; n-- {
					i.Set(int(m), int(n), c)
				}
			} else {
				for n := vStart; n < vStart-length; n++ {
					i.Set(int(m), int(n), c)
				}
			}
		}
	}
	return i
}

func pieChart(size int, values []int64, colorsArray []color.NRGBA, donut bool) image.Image {
	sum := int64(0)
	for _, value := range values {
		sum += value
	}
	i := image.NewNRGBA(image.Rect(0, 0, size, size))
	angle := float64(0)
	r := float64(size / 2.0)
	innerR := 0.4 * r
	for nr, value := range values {
		oldAngle := angle
		angle += 2 * math.Pi * float64(value) / float64(sum)
		c := color.NRGBA{}
		if len(colorsArray) > 0 {
			c = colorsArray[nr%len(colorsArray)]
		} else {
			c = getRandomColor()
		}
		m1, m2, n1, n2 := getSectorBoundingBox(oldAngle, angle, int64(size))
		for m := m1; m <= m2; m++ {
			x := float64(m) - r
			x2 := math.Pow(x, 2.0)
			for n := n1; n <= n2; n++ {
				y := r - float64(n)
				vr := math.Sqrt(x2 + math.Pow(y, 2.0))
				if (innerR < vr || !donut) && vr <= r {
					atan2 := math.Atan2(y, x)
					if atan2 < 0 {
						atan2 = 2*math.Pi + atan2
					}
					if atan2 < angle && atan2 >= oldAngle {
						i.Set(int(m), int(n), c)
					}
				}
			}
		}
	}
	return i
}

func lineChart(size int, values []int64, c color.NRGBA) image.Image {
	min, max := extremes(values)
	i := image.NewNRGBA(image.Rect(0, 0, size, size))
	thickness := float64(size) / float64(len(values))
	span := max - min
	x1 := thickness / 2
	for nr := 1; nr < len(values); nr++ {
		x2 := x1 + thickness
		y1 := float64(size) - float64(values[nr-1]-min)/float64(span)*float64(size)
		y2 := float64(size) - float64(values[nr]-min)/float64(span)*float64(size)
		drawLine(i, c, x1, x2, y1, y2)
		x1 = x2
	}
	return i
}
