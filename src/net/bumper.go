package net

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/db"
	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/hibernate"
	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/marathon"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

//go:generate sh ./../../scripts/gen_static.sh content.go net ./../../web

// Bumper model
type Bumper struct {
	Listener        string
	HostModel       *db.HostEntryManager
	MarathonManager *marathon.Manager
	State           *hibernate.State
}

func (b Bumper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// internal
	if strings.HasPrefix(r.URL.Path, "/-") {
		path := strings.TrimPrefix(r.URL.Path, "/-")
		// hit static non templated content
		if v, ok := Content.PathMapper[path]; ok && v.Template == nil {
			w.Header().Add("Content-Type", v.ContentType)
			w.Write(v.Buffer)
			return
		}
	}

	u, err := url.Parse(r.Referer())
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	ctxData := u.Host
	b.State.SignOfLife([]string{ctxData}, b.HostModel)

	// index
	v := Content.PathMapper["/index.html"]
	w.Header().Add("Content-Type", v.ContentType)
	v.Template.Execute(w, ctxData)
}

func (b *Bumper) getIndexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	path := strings.TrimPrefix(r.URL.Path, "/-")

	// hit content
	if path == "/" {
		if v, ok := Content.PathMapper["/index.html"]; ok {
			w.Header().Add("Content-Type", v.ContentType)
			v.Template.Execute(w, nil)
			return
		}
	}
}

func (b *Bumper) collectActivityHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		http.Error(w, "invalid post data", 400)
	}
	defer r.Body.Close()

	var hostList []string
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&hostList); err != nil {
		log.Error(err)
	}
	log.Infof("SignOfLife %s#v", hostList)

	b.State.SignOfLife(hostList, b.HostModel)

	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte("{}"))
}

func (b *Bumper) getHostHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hostList := []string{}
	for k := range b.State.HostLookup {
		hostList = append(hostList, k)
	}

	data, _ := json.Marshal(hostList)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (b *Bumper) getHostStateHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	buffer, err := base64.StdEncoding.DecodeString(r.URL.Query().Get("host"))

	if err != nil {
		http.NotFound(w, r)
		return
	}

	gID, ok := b.State.HostLookup[string(buffer)]
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

// Run bumper web service
func (b Bumper) Run() error {
	router := httprouter.New()
	router.GET("/-/", b.getIndexHandler)
	router.GET("/-/api/state", b.getHostStateHandler)
	router.GET("/-/api/trigger", b.getHostHandler)
	router.POST("/-/api/trigger", b.collectActivityHandler)
	router.NotFound = b

	return http.ListenAndServe(b.Listener, router)
}
