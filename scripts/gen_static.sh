#!/bin/sh

FILE=$1
PACKAGE=$2
CONTENT=$3

echo "Generate static web content"

FILES=`find ${CONTENT} -type f | grep '^.*\(html\|js\)'`

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
    
${MAPDEF}
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
EOF