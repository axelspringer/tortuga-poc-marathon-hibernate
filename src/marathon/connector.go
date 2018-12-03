package marathon

import (
	"fmt"

	gomarathon "github.com/gambol99/go-marathon"
)

// ConnectorConfig model
type ConnectorConfig struct {
	Endpoint string
}

// Connector model
type Connector struct {
	client gomarathon.Marathon
}

// NewConnector creates a new marathon connector
func NewConnector(c ConnectorConfig) (*Connector, error) {
	connector := &Connector{}

	config := gomarathon.NewDefaultConfig()
	config.URL = c.Endpoint

	client, err := gomarathon.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("marathon connector: %s", err)
	}

	connector.client = client

	return connector, nil
}
