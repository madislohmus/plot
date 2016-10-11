package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	defaultSize = 512
	maxSize     = 1024
	maxValues   = 100

	getPlotDataSQL = `
	SELECT data
	FROM plots
	WHERE url = $1`

	insertPlotDataSQL = `
	INSERT INTO plots
	(url, data)
	VALUES
	($1, $2)`
)

var (
	colorRegex    = regexp.MustCompile("[a-f0-9]{6}")
	templateDir   = flag.String("d", "", "Template directory")
	indexTemplate *template.Template
	db            *sqlx.DB
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
				return clientError{Message: fmt.Sprintf("Maximum allowed size is %d\n", maxSize)}
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
				return nil, clientError{Message: fmt.Sprintf("Could not convert %s to integer", valueString)}
			}
			if value < 0 && !allowNegative {
				return nil, clientError{Message: "Only positive integers allowed"}
			}
			values = append(values, value)
		}
	}
	if len(values) == 0 {
		return nil, clientError{Message: "No values provided"}
	}
	if len(values) > maxValues {
		return nil, clientError{Message: "Maximum 100 values allowed"}
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
				err = clientError{Message: fmt.Sprintf("Color %s at position %d is not a valid html color", colorString, nr)}
				return
			}
			colors = append(colors, getColorFromHTML(colorString))
		}
	}
	return
}

func writeResult(w http.ResponseWriter, buf []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.Write(buf)
}

func getExistingData(r *http.Request) ([]byte, error) {
	var data []byte
	err := db.Get(&data, getPlotDataSQL, r.URL.Path+"?"+r.URL.RawQuery)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	return data, err
}

func insertPlotData(r *http.Request, data []byte) error {
	_, err := db.Exec(insertPlotDataSQL, r.URL.Path+"?"+r.URL.RawQuery, data)
	return err
}

func handleBarChart(w http.ResponseWriter, r *http.Request) {
	data, err := getExistingData(r)
	if err != nil {
		writeError(err, w)
		return
	}
	if data != nil {
		writeResult(w, data)
		logRequest(r)
	}
	sizeInt := defaultSize
	err = getSize(r, &sizeInt)
	if err != nil {
		writeError(err, w)
		return
	}
	values, err := getValues(r, true)
	if err != nil {
		writeError(err, w)
		return
	}
	colors, err := getColors(r)
	if err != nil {
		writeError(err, w)
		return
	}
	buf, err := getBarChartBuffer(sizeInt, values, colors)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not load image"))
		return
	}
	err = insertPlotData(r, buf.Bytes())
	if err != nil {
		writeError(err, w)
		return
	}
	writeResult(w, buf.Bytes())
	logRequest(r)
}

func handlePieChart(w http.ResponseWriter, r *http.Request) {
	data, err := getExistingData(r)
	if err != nil {
		writeError(err, w)
		return
	}
	if data != nil {
		writeResult(w, data)
		logRequest(r)
	}
	sizeInt := defaultSize
	donut := r.FormValue("donut")
	err = getSize(r, &sizeInt)
	if err != nil {
		writeError(err, w)
		return
	}
	values, err := getValues(r, false)
	if err != nil {
		writeError(err, w)
		return
	}
	colors, err := getColors(r)
	if err != nil {
		writeError(err, w)
		return
	}
	buf, err := getPieChartBuffer(sizeInt, values, colors, donut == "true")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not load image"))
		return
	}
	err = insertPlotData(r, buf.Bytes())
	if err != nil {
		writeError(err, w)
		return
	}
	writeResult(w, buf.Bytes())
	logRequest(r)
}

func handleLineChart(w http.ResponseWriter, r *http.Request) {
	data, err := getExistingData(r)
	if err != nil {
		writeError(err, w)
		return
	}
	if data != nil {
		writeResult(w, data)
		logRequest(r)
	}
	sizeInt := defaultSize
	colorString := r.FormValue("c")
	if len(colorString) == 0 {
		colorString = r.FormValue("color")
	}
	err = getSize(r, &sizeInt)
	if err != nil {
		writeError(err, w)
		return
	}
	values, err := getValues(r, true)
	if err != nil {
		writeError(err, w)
		return
	}
	if len(colorString) > 0 {
		matches := colorRegex.MatchString(colorString)
		if !matches {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Color %s is not a valid html color", colorString)))
			return
		}
	}
	buf, err := getLineChartBuffer(sizeInt, values, getColorFromHTML(colorString))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not load image"))
		return
	}
	err = insertPlotData(r, buf.Bytes())
	if err != nil {
		writeError(err, w)
		return
	}
	writeResult(w, buf.Bytes())
	logRequest(r)
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	indexTemplate.Execute(w, map[string]string{"host": r.Host})
	logRequest(r)
}

func initTemplate() {
	var err error
	indexTemplate, err = template.ParseFiles(filepath.Join(*templateDir, "index.html"))
	if err != nil {
		panic(err)
	}
}

func connectToDB() {
	var err error
	dbConnString := os.Getenv("DB_CONN_STRING")
	db, err = sqlx.Connect("postgres", dbConnString)
	if err != nil {
		log.Fatal(err)
	}
}

func logRequest(r *http.Request) {
	_, err := db.Exec("INSERT INTO access (data) VALUES ($1)", fmt.Sprintf("%s %s %s", r.RemoteAddr, r.Method, r.URL))
	if err != nil {
		panic(err)
	}
}

func main() {
	rand.Seed(1) //let's have the same colors
	if !flag.Parsed() {
		flag.Parse()
	}
	initTemplate()
	connectToDB()
	http.HandleFunc("/", handleInfo)
	http.HandleFunc("/bar", handleBarChart)
	http.HandleFunc("/pie", handlePieChart)
	http.HandleFunc("/line", handleLineChart)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
