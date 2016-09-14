package main

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultSize = 512
	maxSize     = 1024
	maxValues   = 100
)

var (
	colorRegex = regexp.MustCompile("[a-f0-9]{6}")
)

func getSize(r *http.Request, target *int) error {
	size := r.FormValue("s")
	if len(size) == 0 {
		size = r.FormValue("size")
	}
	if len(size) > 0 {
		sizei, err := strconv.ParseInt(size, 10, 64)
		if err == nil && sizei > 0 {
			if sizei > maxSize {
				return fmt.Errorf("Maximum allowed size is %d\n", maxSize)
			}
			*target = int(sizei)
		}
	}
	return nil
}

func getValues(r *http.Request, allowNegative bool) (values []int64, err error) {
	valuesString := r.FormValue("v")
	if len(valuesString) == 0 {
		valuesString = r.FormValue("values")
	}
	if len(valuesString) > 0 {
		valuesStringArray := strings.Split(valuesString, ",")
		for _, valueString := range valuesStringArray {
			value, err := strconv.ParseInt(valueString, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("Could not convert %s to integer", valueString)
			}
			if value < 0 && !allowNegative {
				return nil, fmt.Errorf("Only positive integers allowed")
			}
			values = append(values, value)
		}
	}
	if len(values) == 0 {
		return nil, errors.New("No values provided")
	}
	if len(values) > maxValues {
		return nil, errors.New("Maximum 100 values allowed")
	}
	return
}

func getColors(r *http.Request) (colors []color.NRGBA, err error) {
	colorsString := r.FormValue("c")
	if len(colorsString) == 0 {
		colorsString = r.FormValue("colors")
	}
	if len(colorsString) > 0 {
		colorsStringArray := strings.Split(colorsString, ",")
		for nr, colorString := range colorsStringArray {
			matches := colorRegex.MatchString(colorString)
			if !matches {
				err = fmt.Errorf("Color %s at position %d is not a valid html color", colorString, nr)
				return
			}
			colors = append(colors, getColorFromHTML(colorString))
		}
	}
	return
}

func writeResult(w http.ResponseWriter, buf *bytes.Buffer) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))
	w.Write(buf.Bytes())
}

func handleBarChart(w http.ResponseWriter, r *http.Request) {
	sizeInt := defaultSize
	err := getSize(r, &sizeInt)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	values, err := getValues(r, true)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	colors, err := getColors(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	buf, err := getBarChartBuffer(sizeInt, values, colors)
	if err != nil {
		w.Write([]byte("Could not load image"))
		return
	}
	writeResult(w, buf)
}

func handlePieChart(w http.ResponseWriter, r *http.Request) {
	sizeInt := defaultSize
	donut := r.FormValue("donut")
	err := getSize(r, &sizeInt)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	values, err := getValues(r, false)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	colors, err := getColors(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	buf, err := getPieChartBuffer(sizeInt, values, colors, donut == "true")
	if err != nil {
		w.Write([]byte("Could not load image"))
		return
	}
	writeResult(w, buf)
}

func handleLineChart(w http.ResponseWriter, r *http.Request) {
	sizeInt := defaultSize
	colorString := r.FormValue("c")
	if len(colorString) == 0 {
		colorString = r.FormValue("color")
	}
	err := getSize(r, &sizeInt)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	values, err := getValues(r, true)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	if len(colorString) > 0 {
		matches := colorRegex.MatchString(colorString)
		if !matches {
			w.Write([]byte(fmt.Sprintf("Color %s is not a valid html color", colorString)))
			return
		}
	}
	buf, err := getLineChartBuffer(sizeInt, values, getColorFromHTML(colorString))
	if err != nil {
		w.Write([]byte("Could not load image"))
		return
	}
	writeResult(w, buf)
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html><head></head><body><p>Examples</p><a href="http://plotme.ee/bar?size=100&values=1,2,3&colors=ff0000,00ff00,0000ff">bar</a><br><a href="http://plotme.ee/line?size=500&values=1,2,1&color=ff0000">line</a><br><a href="http://plotme.ee/pie?size=500&values=1,2,4&colors=ff0000,00ff00,0000ff">pie</a>`))
}

func main() {
	rand.Seed(time.Now().Unix())
	http.HandleFunc("/", handleInfo)
	http.HandleFunc("/bar", handleBarChart)
	http.HandleFunc("/pie", handlePieChart)
	http.HandleFunc("/line", handleLineChart)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
