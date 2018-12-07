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
	log.Infof("fallback ServeHTTP r.URL %#v", r.URL)

	// internal
	if strings.HasPrefix(r.URL.Path, "/-") {
		path := strings.TrimPrefix(r.URL.Path, "/-")
		log.Infof("Serve static files %s", path)

		// hit static non templated content
		if v, ok := Content.PathMapper[path]; ok && v.Template == nil {
			w.Header().Add("Content-Type", v.ContentType)
			w.Write(v.Buffer)
			return
		}
	}

	b.redirectHandler(w, r)
}

func (b *Bumper) getIndexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	path := strings.TrimPrefix(r.URL.Path, "/-")

	log.Infof("indexHandler r.URL %#v path %s", r.URL, path)

	// hit content
	if path == "/" {
		if v, ok := Content.PathMapper["/index.html"]; ok {
			w.Header().Add("Content-Type", v.ContentType)
			v.Template.Execute(w, nil)

			// heartbeat
			//h := r.Host
			//b.State.SignOfLife([]string{h}, b.HostModel)
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

	log.Infof("collectActivityHandler %s", string(body))

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

func (b *Bumper) redirectHandler(w http.ResponseWriter, r *http.Request) {
	referer := r.Referer()
	u, err := url.Parse(referer)

	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), 400)
		return
	}

	redirectURL := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		User:   u.User,
		Path:   "/-/",
	}

	qs := redirectURL.Query()
	qs.Set("t", base64.StdEncoding.EncodeToString([]byte(referer)))
	redirectURL.RawQuery = qs.Encode()

	log.Infof("redirectHandler redirect %s", redirectURL.String())

	http.Redirect(w, r, redirectURL.String(), 307)
}

// Run bumper web service
func (b Bumper) Run() error {
	router := httprouter.New()
	router.GET("/-/", b.getIndexHandler)
	router.GET("/-/api/trigger", b.getHostHandler)
	router.POST("/-/api/trigger", b.collectActivityHandler)
	router.NotFound = b

	return http.ListenAndServe(b.Listener, router)
}
