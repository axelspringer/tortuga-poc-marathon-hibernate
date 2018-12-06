package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

const (
	envPrefix             = "HIBERTHON_"
	defaultCollectionTime = 10
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

/*

* -webhook *url* Webhook for the hiberthon api (Required)
* -logfile *path* Path to the logfile to watch (Required)
* -format *fmt* like "clf" (Required)
* -collection-time *time* in seconds (default 10s)
 */

// Configuration model
type Configuration struct {
	webhookURL     string
	logfilePath    string
	format         string
	collectionTime int
}

func (c *Configuration) Read() {
	webhookURLPtr := getFlagPtr("webhook", "Webhook for the hiberthon api (Required)")
	logfilePathPtr := getFlagPtr("logfile", "Path to the logfile to watch (Required)")
	logFormatPtr := getFlagPtr("format", "Logfile like \"clf\" (Required)")
	collectionTimePtr := getFlagPtr("collection-time", "in seconds (default 10s)")

	flag.Parse()

	if webhookURLPtr != nil && *webhookURLPtr != "" {
		c.webhookURL = *webhookURLPtr
	}

	if logfilePathPtr != nil && *logfilePathPtr != "" {
		c.logfilePath = *logfilePathPtr
	}

	if logFormatPtr != nil && *logFormatPtr != "" {
		c.format = *logFormatPtr
	}

	c.collectionTime = defaultCollectionTime
	if collectionTimePtr != nil && *collectionTimePtr != "" {
		ct := *collectionTimePtr
		c.collectionTime, _ = strconv.Atoi(ct)
		if c.collectionTime == 0 {
			c.collectionTime = defaultCollectionTime
		}
	}
}
