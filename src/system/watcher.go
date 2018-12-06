package system

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// FileWatcher model
type FileWatcher struct {
	FileName           string
	OldFilePosition    int64
	ProcessHandler     func([]byte)
	HostCollectHandler func(map[string]int)
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
		FileName:           filePath,
		OldFilePosition:    0,
		HostCollectHandler: func(map[string]int) {}, // nop
		ProcessHandler:     func([]byte) {},         // nop
	}

	switch true {
	case strings.HasPrefix(format, "regex:"):
		log.Info("set regex process handler")
		fw.ProcessHandler = regexProcessHandler(format, fw)
	case strings.HasPrefix(format, "seq:"):
		log.Info("set seq process handler")
		fw.ProcessHandler = seqProcessHandler(format, fw)
	}

	return fw
}

func seqProcessHandler(format string, fw *FileWatcher) func([]byte) {
	seqFmt := format[len("seq:"):]

	params := strings.Split(seqFmt, ":")
	if len(params) < 2 {
		log.Panicf("Can't create sequence: %s", format)
	}

	separator := params[0]
	index, _ := strconv.Atoi(params[1])

	return func(buffer []byte) {
		hostMap := map[string]int{}
		scanner := bufio.NewScanner(strings.NewReader(string(buffer)))

		for scanner.Scan() {
			lineSplit := strings.Split(scanner.Text(), separator)
			if len(lineSplit) > index {
				h := lineSplit[index]
				switch hm, ok := hostMap[h]; ok {
				case true:
					hostMap[h] = hm + 1
				case false:
					hostMap[h] = 1
				}
			}
		}

		log.Infof("hostmap %#v", hostMap)
		fw.HostCollectHandler(hostMap)
	}
}

func regexProcessHandler(format string, fw *FileWatcher) func([]byte) {
	regexStr := format[len("regex:"):]
	regex, err := regexp.Compile(regexStr)

	if err != nil {
		log.Panicf("Can't compile regex: %s", err)
	}

	return func(buffer []byte) {
		findings := regex.FindAllString(string(buffer), -1)
		hostMap := map[string]int{}
		for _, match := range findings {
			switch hm, ok := hostMap[match]; ok {
			case true:
				hostMap[match] = hm + 1
			case false:
				hostMap[match] = 1
			}
		}
		log.Infof("hostmap %#v", hostMap)
		fw.HostCollectHandler(hostMap)
	}
}
