package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	mux        = http.NewServeMux()
	listenAddr string
)

type (
	Content struct {
		Title    string
		Hostname string
		Extended Extended
	}
)

// see sample/test-data.json for example that supports this struct
type Extended struct {
	Data []ExtendedData
}
type ExtendedData struct {
	Key   string
	Value string
}

func (ext *Extended) KeyValue(k string) string {
	for i := range ext.Data {
		extdata := ext.Data[i]
		if extdata.Key == k {
			return extdata.Value
		}
	}
	return ""
}

func init() {
	flag.StringVar(&listenAddr, "listen", ":8080", "listen address")
}

func loadTemplate(filename string) (*template.Template, error) {
	return template.ParseFiles(filename)
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Printf("request from %s\n", r.Header.Get("X-Forwarded-For"))
	t, err := loadTemplate("templates/index.html.tmpl")
	if err != nil {
		fmt.Printf("error loading template: %s\n", err)
		return
	}

	title := os.Getenv("TITLE")

	// lets read some json from DATAFILE_EXT so that we can include it in the template
	// example, this is going to be useful for testing data volumes for example
	var extended Extended
	os_datafile := os.Getenv("DATAFILE_EXT")
	if os_datafile != "" {
		// something was passed, lets read it
		fmt.Printf("loading additional data from: %s\n", os_datafile)
		extendedfile, e := ioutil.ReadFile(os_datafile)
		if e != nil {
			fmt.Printf("File error: %s\n", err)
			return
		}
		err := json.Unmarshal(extendedfile, &extended)
		if err != nil {
			fmt.Printf("error loading json: %s\n", err)
			return
		}
	} else {
		fmt.Printf("skipping additional data, no file found in env DATAFILE_EXT\n")
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	cnt := &Content{
		Title:    title,
		Hostname: hostname,
		Extended: extended,
	}

	t.Execute(w, cnt)
}

func ping(w http.ResponseWriter, r *http.Request) {
	resp := fmt.Sprintf("ehazlett/docker-demo: hello %s\n", r.RemoteAddr)
	w.Write([]byte(resp))
}

func main() {
	flag.Parse()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/ping", ping)
	mux.HandleFunc("/", index)

	log.Printf("listening on %s\n", listenAddr)

	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatalf("error serving: %s", err)
	}
}
