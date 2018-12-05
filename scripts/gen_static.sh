#!/bin/sh

FILE=$1
PACKAGE=$2
CONTENT=$3

echo "Generate static web content"

FILES=`find ${CONTENT} -type f | grep '^.*\(html\|js\|svg\|css\)'`

MAPDEF=""

for line in ${FILES}; do
    data=$(base64 "${line}")
    path="${line}"
    path=$(echo "${path}" | sed -e 's/[\.\/]*web//g') 
    echo "add file ${line} with path ${path} to content mapper"
    MAPDEF="${MAPDEF}    Content.Add(\"${path}\", \"${data}\")"$'\n'
done

cat << EOF > ${FILE}
package ${PACKAGE}

// DO NOT EDIT. THIS FILE IS GENERATED

import (
	"encoding/base64"
	"log"
	"path/filepath"
	"strings"
	"text/template"
)

// WebEntry single entry
type WebEntry struct {
	// ContentType
	ContentType string
	// Buffer of static file
	Buffer []byte
    // Template
	Template *template.Template
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

    var t *template.Template

    contentType := "text/plain"
	switch ext := strings.ToLower(filepath.Ext(p)); ext {
    case ".svg", ".svgz":
        contentType = "image/svg+xml"
	case ".css":
		contentType = "text/css"
    case ".png":
		contentType = "application/png"
	case ".html":
        contentType = "text/html"
		t = sw.Template.New(p)
		if _, err := t.Parse(string(decoded)); err != nil {
			log.Fatal(err)
		}
	case ".js":
		contentType = "text/javascript"
	}

	sw.PathMapper[p] = WebEntry{
		Buffer:      decoded,
		ContentType: contentType,
        Template:    t,
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
    
${MAPDEF}
}
EOF