package main

import (
	"time"

	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/db"
	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/hibernate"
	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/marathon"
	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/net"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Println("hiberthon")

	config := &Configuration{}
	config.Read()

	// setup db connection
	dbConnector, err := db.NewConnector(db.ConnectorConfig{
		Endpoint:  config.dbEndpoint,
		Key:       config.dbKey,
		Secret:    config.dbSecret,
		TableName: config.dbTablePrefix,
		Region:    config.dbRegion,
	})

	if err != nil {
		log.Fatal(err)
	}

	// setup marathon connection
	marathonConnector, err := marathon.NewConnector(marathon.ConnectorConfig{
		Endpoint: config.marathonEndpoint,
	})

	if err != nil {
		log.Fatal(err)
	}

	state := hibernate.NewState()

	// web bumper
	bumper := &net.Bumper{
		Listener:        config.httpListener,
		HostModel:       db.NewHostEntryManager(dbConnector),
		MarathonManager: marathon.NewManager(marathonConnector),
		State:           state,
	}

	ticker := time.NewTicker(time.Duration(config.stateUpdateTime) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				entryList := bumper.HostModel.GetAllEntries()
				appList := bumper.MarathonManager.ListAppsWithHibernationSupport()
				state.Update(bumper.HostModel, bumper.MarathonManager, entryList, appList)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	log.Fatal(bumper.Run())
}
