package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

func generate(w io.Writer) error {
	data, err := Asset("data/template.html")
	if err != nil {
		return fmt.Errorf("asset not found")
	}

	tmpl, err := template.New("index").Parse(string(data))
	if err != nil {
		return err
	}

	templates := make(map[string]string)

	paths, err := filepath.Glob("gene-attributes/templates/*")

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		t, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		templates[filepath.Base(path)] = string(t)
	}

	buffer, err := json.Marshal(templates)
	if err != nil {
		return err
	}

	f, err := os.Open("gene-attributes/index.js")
	if err != nil {
		return err
	}
	defer f.Close()

	js, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	b := struct {
		TemplatesJson string
		IndexJs       string
		ElementName   string
	}{
		TemplatesJson: string(buffer),
		IndexJs:       string(js),
		ElementName:   "togostanza-gene-attributes",
	}

	return tmpl.Execute(w, b)
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/gene-attributes/", func(w http.ResponseWriter, req *http.Request) {
		err := generate(w)
		if err != nil {
			log.Println("ERROR", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	mux.Handle("/", http.FileServer(http.Dir(".")))

	port := 8080
	addr := fmt.Sprintf(":%d", port)
	log.Println("listening on", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
