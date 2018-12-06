package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/system"
	log "github.com/sirupsen/logrus"
)

var (
	watcher *system.FileWatcher
	client  = &http.Client{
		Timeout: 60 * time.Second,
	}
)

func getHostList(endpoint string) ([]string, error) {
	res := []string{}

	r, err := client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&res)
	return res, err
}

func postActiveHosts(endpoint string, hosts []string) error {
	if hosts == nil || len(hosts) == 0 {
		return nil
	}

	jsonData, err := json.Marshal(hosts)
	if err != nil {
		return err
	}

	if _, err := client.Post(endpoint, "application/json", bytes.NewReader(jsonData)); err != nil {
		return err
	}

	return nil
}

func hostUpdate(endpointURL string, watcher *system.FileWatcher) error {
	hostList, err := getHostList(endpointURL)
	if err != nil {
		log.Errorf("Error while updating hostlist. %s", err)
		return err
	}

	watcher.SetHosts(hostList)

	err = postActiveHosts(endpointURL, watcher.GetActiveHostList())
	if err != nil {
		log.Errorf("Error while posting active hosts. %s", err)
		return err
	}

	return nil
}

func logUpdate(endpointURL string, watcher *system.FileWatcher) error {
	err := watcher.Process()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.SetLevel(log.DebugLevel)

	log.Info("hiberthon trigger")
	// read config
	config := &Configuration{}
	log.Info("hiberthon config read")
	config.Read()
	// debug output
	log.Debugf("using logfile %s", config.logfilePath)
	log.Debugf("host update interval %d seconds", config.hostUpdateTime)
	log.Debugf("collector interval %d seconds", config.collectionTime)
	log.Debugf("hiberthon API endpoint %s", config.endpointURL)
	log.Debugf("parser format %s", config.format)
	// watcher init
	log.Info("new watcher")
	watcher = system.NewFileWatcher(config.logfilePath, config.format)
	// interval ticker
	logUpdateTicker := time.NewTicker(time.Duration(config.collectionTime) * time.Second)
	hostUpdateTicker := time.NewTicker(time.Duration(config.hostUpdateTime) * time.Second)
	// interrupt chan
	quit := make(chan struct{})
	// set the host list
	if err := hostUpdate(config.endpointURL, watcher); err != nil {
		log.Error(err)
	}
	// update host collector map
	go func() {
		for _ = range hostUpdateTicker.C {
			if err := hostUpdate(config.endpointURL, watcher); err != nil {
				log.Error(err)
			}
		}
	}()
	// read host activity
	go func() {
		for _ = range logUpdateTicker.C {
			if err := logUpdate(config.endpointURL, watcher); err != nil {
				log.Error(err)
			}
		}
	}()
	// wait forever
	<-quit
}
