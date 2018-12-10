package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

const (
	envPrefix              = "HIBERTHON_"
	defaultStateUpdateTime = 10
)

// getOSPrefixEnv get os env
func getOSPrefixEnv(s string) string {
	s = strings.ToUpper(strings.Replace(s, "-", "_", -1))

	if e := strings.TrimSpace(os.Getenv(envPrefix + s)); len(e) > 0 {
		return e
	}

	return ""
}

// getFlagPtr
func getFlagPtr(name string, doc string) *string {
	return flag.String(name, getOSPrefixEnv(name), doc)
}

// Configuration model
type Configuration struct {
	dbEndpoint       string
	dbRegion         string
	dbKey            string
	dbSecret         string
	dbTablePrefix    string
	marathonEndpoint string
	httpListener     string
	stateUpdateTime  int
}

func (c *Configuration) Read() {
	dbEndpointPtr := getFlagPtr("db-endpoint", "DynamoDB endpoint (Required)")
	dbRegionPtr := getFlagPtr("db-region", "DynamoDB region (Required)")
	dbKeyPtr := getFlagPtr("db-key", "DynamoDB credential key")
	dbSecretPtr := getFlagPtr("db-secret", "DynamoDB credential secret")
	dbTablePrefixPtr := getFlagPtr("db-table-prefix", "DynamoDB table name prefix")
	marathonEndpointPtr := getFlagPtr("marathon-endpoint", "DynamoDB endpoint (Required)")
	httpListenerPtr := getFlagPtr("listener", "Web listener (Required)")
	stateUpdateTimePtr := getFlagPtr("state-update", "State update interval tim in seconds (default 10)")

	flag.Parse()

	if dbEndpointPtr != nil && *dbEndpointPtr != "" {
		c.dbEndpoint = *dbEndpointPtr
	}

	if dbTablePrefixPtr != nil && *dbTablePrefixPtr != "" {
		c.dbTablePrefix = *dbTablePrefixPtr
	}

	if dbRegionPtr != nil && *dbRegionPtr != "" {
		c.dbRegion = *dbRegionPtr
	}

	if dbKeyPtr != nil && dbSecretPtr != nil && *dbKeyPtr != "" && *dbSecretPtr != "" {
		c.dbKey = *dbKeyPtr
		c.dbSecret = *dbSecretPtr
	}

	if marathonEndpointPtr != nil && *marathonEndpointPtr != "" {
		c.marathonEndpoint = *marathonEndpointPtr
	}

	if httpListenerPtr != nil && *httpListenerPtr != "" {
		c.httpListener = *httpListenerPtr
	}

	c.stateUpdateTime = defaultStateUpdateTime
	if stateUpdateTimePtr != nil && *stateUpdateTimePtr != "" {
		ct := *stateUpdateTimePtr
		c.stateUpdateTime, _ = strconv.Atoi(ct)
		if c.stateUpdateTime == 0 {
			c.stateUpdateTime = defaultStateUpdateTime
		}
	}
}
