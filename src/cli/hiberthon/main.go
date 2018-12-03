package main

import (
	"fmt"

	"github.com/axelspringer/poc-marathon-hibernate/src/net"

	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("hiberthon")

	config := &Configuration{}
	config.Read()
	/*
		// setup db connection
		dbConnector, err := db.NewConnector(db.ConnectorConfig{
			Endpoint:  config.dbEndpoint,
			Key:       config.dbKey,
			Secret:    config.dbSecret,
			TableName: config.dbTablePrefix,
			Region:    config.dbRegion,
		})

		if err != nil {
			logrus.Fatal(err)
		}

		// setup marathon connection
		marathonConnector, err := marathon.NewConnector(marathon.ConnectorConfig{
			Endpoint: config.marathonEndpoint,
		})

		if err != nil {
			logrus.Fatal(err)
		}
	*/
	// web bumper
	bumper := &net.Bumper{
		Listener: config.httpListener,
	}
	logrus.Fatal(bumper.Run())
}
