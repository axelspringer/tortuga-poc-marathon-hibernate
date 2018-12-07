package system

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// FileWatcher model
type FileWatcher struct {
	sync.RWMutex
	HostCollector   map[string]int
	FileName        string
	OldFilePosition int64
	ProcessHandler  func([]byte)
}

// SetHosts sets the collector hostmap
func (fw *FileWatcher) SetHosts(hosts []string) {
	hostMap := map[string]int{}

	for _, h := range hosts {
		c := 0
		if v, ok := fw.HostCollector[h]; ok {
			c = v
		}
		hostMap[h] = c
	}

	fw.Lock()
	defer fw.Unlock()
	fw.HostCollector = hostMap
}

// Collect discrete host informations
func (fw *FileWatcher) Collect(discreteHM map[string]int) {
	fw.Lock()
	defer fw.Unlock()

	for k, v := range fw.HostCollector {
		if count, ok := discreteHM[k]; ok {
			fw.HostCollector[k] = v + count
		}
	}
}

// GetActiveHostList returns a list of active hosts
func (fw *FileWatcher) GetActiveHostList() []string {
	fw.Lock()
	defer fw.Unlock()

	resList := []string{}

	for k, v := range fw.HostCollector {
		if v > 0 {
			resList = append(resList, k)
		}
		// reset collector
		fw.HostCollector[k] = 0
	}

	return resList
}

// Process the logfile trigger
func (fw *FileWatcher) Process() error {
	st, err := os.Stat(fw.FileName)
	if err != nil {
		return err
	}
	// get the size
	size := st.Size()
	// reset old position if file was renewed
	if size < fw.OldFilePosition {
		fw.OldFilePosition = 0
	}
	// read diff
	if size > fw.OldFilePosition {
		fh, err := os.Open(fw.FileName)
		if err != nil {
			return err
		}
		defer fh.Close()

		fh.Seek(fw.OldFilePosition, 0)
		diff := make([]byte, size-fw.OldFilePosition)
		fh.Read(diff)

		fw.OldFilePosition = size
		fw.ProcessHandler(diff)
	}

	return nil
}

// NewFileWatcher creates a new instance
func NewFileWatcher(filePath string, format string) *FileWatcher {
	log.Infof("new file watcher fmt: %s", format)

	fw := &FileWatcher{
		FileName:        filePath,
		OldFilePosition: 0,
		ProcessHandler:  func([]byte) {}, // nop
		HostCollector:   map[string]int{},
	}

	switch true {
	case strings.HasPrefix(format, "traefik:clf"):
		log.Info("set traefik clf process handler")
		fw.ProcessHandler = traefikCLFProcessHandler(format, fw)
	}
	// set the initial offset
	if st, err := os.Stat(fw.FileName); err == nil {
		fw.OldFilePosition = st.Size()
	}
	
	return fw
}

func traefikCLFProcessHandler(format string, fw *FileWatcher) func([]byte) {
	regex, err := regexp.Compile(`[^\s"']+|"([^"]*)"|'([^']*)`)

	if err != nil {
		log.Panicf("Can't compile regex: %s", err)
	}

	return func(buffer []byte) {
		hostMap := map[string]int{}
		scanner := bufio.NewScanner(strings.NewReader(string(buffer)))

		for scanner.Scan() {
			findings := regex.FindAllString(string(scanner.Text()), -1)

			if len(findings) < 11 {
				continue
			}

			if strings.HasPrefix(findings[4], "+") || strings.HasPrefix(findings[4], "-") {
				tzOffset := findings[4]
				findings = append(findings[:4], findings[5:]...)
				findings[3] = findings[3] + " " + tzOffset
			}

			pathSplit := strings.Split(findings[4], " ")
			if len(pathSplit) != 3 {
				log.Errorf("Unable to split path component %s", findings[4])
				continue
			}

			path := pathSplit[1]
			// strip querystring
			if i := strings.Index(path, "?"); i != -1 {
				path = path[:i]
			}

			host := strings.Trim(findings[10], "\"")

			// skip observation paths
			trimPath := strings.TrimRight(path, "/")
			if trimPath == "/metrics" || trimPath == "/health" {
				continue
			}

			switch hm, ok := hostMap[host]; ok {
			case true:
				hostMap[host] = hm + 1
			case false:
				hostMap[host] = 1
			}
		}

		fw.Collect(hostMap)
	}
}
