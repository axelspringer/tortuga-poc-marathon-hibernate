package main

import (
	"time"

	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/system"
	log "github.com/sirupsen/logrus"
)

var watcher *system.FileWatcher

func update() {
	log.Info("update")

	err := watcher.Process()
	if err != nil {
		log.Error(err)
	}
}

func main() {
	log.Info("hiberthon trigger")
	// read config
	config := &Configuration{}
	log.Info("hiberthon config read")
	config.Read()
	// watcher init
	log.Info("new watcher")
	watcher = system.NewFileWatcher(config.logfilePath, config.format)
	// interval ticker
	log.Infof("collector interval %d seconds", config.collectionTime)
	ticker := time.NewTicker(time.Duration(config.collectionTime) * time.Second)
	quit := make(chan struct{})

	go func() {
		for _ = range ticker.C {
			update()
		}
	}()

	// wait forever
	<-quit
}
