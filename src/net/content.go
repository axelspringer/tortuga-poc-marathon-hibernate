package net

// DO NOT EDIT. THIS FILE IS GENERATED

import (
	"encoding/base64"
	"log"
	"text/template"
    "strings"
    "path/filepath"
)

// WebEntry single entry
type WebEntry struct {
	// ContentType
	ContentType string
	// Buffer of static file
	Buffer []byte
}

// StaticContent mapper for files
type StaticContent struct {
    PathMapper map[string]WebEntry
	Template   *template.Template
}

// Add a new entry to the path mapper
func (sw *StaticContent) Add(p string, b string) {
	decoded, err := base64.StdEncoding.DecodeString(b)
    if err != nil {
        log.Fatalf("Unable to decode static file")
    }

    contentType := "text/plain"
	switch ext := strings.ToLower(filepath.Ext(p)); ext {
	case ".png":
		contentType = "application/png"
	case ".html":
		contentType = "text/html"
	case ".js":
		contentType = "text/javascript"
	}

	sw.PathMapper[p] = WebEntry{
		Buffer:      decoded,
		ContentType: contentType,
	}
}

// Content accessor
var Content StaticContent

func init() {
    // initialize
    Content = StaticContent{
		PathMapper: map[string]WebEntry{},
        Template:   template.New("root"),
	}
    
    Content.Add("/index.html", "PCFET0NUWVBFIGh0bWw+CjxodG1sIGxhbmc9ImVuIj4KICA8aGVhZD4KICAgIDxtZXRhIGNoYXJzZXQ9InV0Zi04Ij4KICAgIDx0aXRsZT50aXRsZTwvdGl0bGU+CiAgICA8bGluayByZWw9InN0eWxlc2hlZXQiIGhyZWY9Ii9jc3Mvc3R5bGUuY3NzIj4KICA8L2hlYWQ+CiAgPGJvZHk+CiAgICA8aDE+QnVtcGVyPC9oMT4KICAgIDxzY3JpcHQgc3JjPSIvanMvYnVtcGVyLmpzIj48L3NjcmlwdD4KICA8L2JvZHk+CjwvaHRtbD4=")
    Content.Add("/js/bumper.js", "YWxlcnQoInNjcmlwdCBsb2FkZWQiKTs=")

    // generate templates
	for k, v := range Content.PathMapper {
		name := k
        log.Printf("Parsing template %s", name)
		t := Content.Template.New(name)
		if _, err := t.Parse(string(v.Buffer)); err != nil {
			log.Fatal(err)
		}
	}
}
