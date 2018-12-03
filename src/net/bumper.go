package net

import (
	"fmt"
	"net/http"
)

//go:generate sh ./../../scripts/gen_static.sh content.go net ./../../web

// Bumper model
type Bumper struct {
	Listener string
}

func (b Bumper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, you've hit %s\n", r.URL.Path)
}

// Run bumper web service
func (b Bumper) Run() error {
	return http.ListenAndServe(b.Listener, b)
}
