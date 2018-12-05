package net

import (
	"net/http"

	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/db"
	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/hibernate"
	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/marathon"
)

//go:generate sh ./../../scripts/gen_static.sh content.go net ./../../web

// Bumper model
type Bumper struct {
	Listener        string
	HostModel       *db.HostEntryManager
	MarathonManager *marathon.Manager
	State           *hibernate.State
}

func (b Bumper) serveAPIState(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	qsHost := r.URL.Query().Get("host")
	gID, ok := b.State.HostLookup[qsHost]
	if !ok {
		http.NotFound(w, r)
		return
	}

	he, ok := b.State.GroupMap[gID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(he.ToJSON()))
}

func (b Bumper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	he := db.HostEntry{
		State: "hibernate",
	}

	if r.URL.Path == "/api/state" {
		b.serveAPIState(w, r)
		return
	}

	if r.URL.Path == "/api/alive" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(he.ToJSON()))
		return
	}

	// hit static non templated content
	if v, ok := Content.PathMapper[r.URL.Path]; ok && v.Template == nil {
		w.Header().Add("Content-Type", v.ContentType)
		w.Write(v.Buffer)
		return
	}

	// hit content
	if r.URL.Path == "/" || r.URL.Path == "index.html" {
		if v, ok := Content.PathMapper["/index.html"]; ok {
			w.Header().Add("Content-Type", v.ContentType)
			v.Template.Execute(w, he)

			// heartbeat
			h := r.Host
			b.State.Lock()
			defer b.State.Unlock()
			if gID, ok := b.State.HostLookup[h]; ok {
				b.HostModel.UpdateLatestUsage(gID)
			}
			return
		}
	}

	// hit 404
	http.NotFound(w, r)
}

// Run bumper web service
func (b Bumper) Run() error {
	return http.ListenAndServe(b.Listener, b)
}
