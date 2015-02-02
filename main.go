package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type Stanza struct {
	BaseDir string
	Name    string
}

func NewStanza(baseDir, name string) *Stanza {
	st := &Stanza{
		BaseDir: baseDir,
		Name:    name,
	}
	if !st.MetadataExists() {
		return nil
	}
	// TODO: validate metadata

	return st
}

func (st *Stanza) MetadataPath() string {
	return path.Join(st.BaseDir, "metadata.json")
}

func (st *Stanza) MetadataExists() bool {
	_, err := os.Stat(st.MetadataPath())
	if err != nil {
		return false
	}
	return true
}

func (st *Stanza) TemplateGlobPattern() string {
	return path.Join(st.BaseDir, "templates/*")
}

func (st *Stanza) IndexJsPath() string {
	return path.Join(st.BaseDir, "index.js")
}

func (st *Stanza) Generate(w io.Writer) error {
	data, err := Asset("data/template.html")
	if err != nil {
		return fmt.Errorf("asset not found")
	}

	tmpl, err := template.New("index").Parse(string(data))
	if err != nil {
		return err
	}

	templates := make(map[string]string)

	paths, err := filepath.Glob(st.TemplateGlobPattern())

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

	f, err := os.Open(st.IndexJsPath())
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
		ElementName:   "togostanza-" + st.Name,
	}

	return tmpl.Execute(w, b)
}

func main() {
	mux := http.NewServeMux()
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	assetsHandler := http.FileServer(http.Dir(cwd))

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		stanzaName := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")
		st := NewStanza(path.Join(cwd, stanzaName), stanzaName)
		if st == nil {
			assetsHandler.ServeHTTP(w, req)
			return
		}
		err := st.Generate(w)
		if err != nil {
			log.Println("ERROR", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	port := 8080
	addr := fmt.Sprintf(":%d", port)
	log.Println("listening on", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
