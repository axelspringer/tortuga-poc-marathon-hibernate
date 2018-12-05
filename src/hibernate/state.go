package hibernate

import (
	"fmt"
	"sync"
	"time"

	"github.com/axelspringer/tortuga-poc-marathon-hibernate/src/db"
	marathonManager "github.com/axelspringer/tortuga-poc-marathon-hibernate/src/marathon"
	"github.com/gambol99/go-marathon"

	log "github.com/sirupsen/logrus"
)

// AppCluster holds the informations for grouped applications
type AppCluster struct {
	ID           string
	Applications map[string]float32
}

// State represents the global hibernation state
type State struct {
	sync.RWMutex
	HostLookup map[string]string
	GroupMap   map[string]db.HostEntry
}

// NewState creates a new instance
func NewState() *State {
	s := &State{
		HostLookup: map[string]string{},
		GroupMap:   map[string]db.HostEntry{},
	}
	return s
}

func filterAppsByGroup(apps []marathon.Application, groupID string) []marathon.Application {
	resList := []marathon.Application{}

	for _, v := range apps {
		if v.Labels != nil {
			if gID, ok := (*v.Labels)["hiberthon.group"]; ok && groupID == gID {
				resList = append(resList, v)
			}
		}
	}

	return resList
}

func scaleMapDiffers(apps []marathon.Application, hostEntry db.HostEntry) (map[string]int, bool) {
	newScaleMap := map[string]int{}
	isDifferent := false

	for _, v := range apps {
		inst := 0
		if v.Instances != nil {
			inst = *v.Instances
		}
		newScaleMap[v.ID] = inst

		if hostEntry.ScaleMap == nil {
			isDifferent = true
			continue
		}

		if v, ok := hostEntry.ScaleMap[v.ID]; !ok || v != inst {
			isDifferent = true
		}
	}

	isDifferent = isDifferent || len(newScaleMap) != len(hostEntry.ScaleMap)
	return newScaleMap, isDifferent
}

// applyScaleMap merges the scale map
func applyScaleMap(apps []marathon.Application, sm map[string]int) []marathon.Application {
	for index, a := range apps {
		if numInstances, ok := sm[a.ID]; ok {
			instances := numInstances
			apps[index].Instances = &instances
		}
	}

	return apps
}

// Hibernate by host
func (s *State) Hibernate(gID string, hostManager *db.HostEntryManager, m *marathonManager.Manager) error {
	// set transitional state
	hostManager.UpdateState(gID, db.HostStateDozeOff)
	// scale to zero
	zero := 0
	groupAppList := m.ListAppsByLableGroup(gID)
	for index := range groupAppList {
		groupAppList[index].Instances = &zero
	}
	// deploy changes
	m.Deploy(groupAppList)
	hostManager.UpdateState(gID, db.HostStateHibernate)
	return nil
}

// Wakeup by host
func (s *State) Wakeup(gID string, hostManager *db.HostEntryManager, m *marathonManager.Manager) error {
	s.Lock()
	defer s.Unlock()

	gData, ok := s.GroupMap[gID]
	if !ok {
		return fmt.Errorf("Unknown group %s", gID)
	}
	// set transitional state
	hostManager.UpdateState(gID, db.HostStateWakeUp)
	// scale to old values
	groupAppList := m.ListAppsByLableGroup(gID)
	groupAppList = applyScaleMap(groupAppList, gData.ScaleMap)
	// deploy changes
	m.Deploy(groupAppList)
	hostManager.UpdateState(gID, db.HostStateRun)
	return nil
}

// Update the global state
func (s *State) Update(hostManager *db.HostEntryManager, m *marathonManager.Manager, entries []db.HostEntry, apps []marathon.Application) {
	currentTS := time.Now().Unix()
	groupMap := map[string]db.HostEntry{}
	hostLookup := map[string]string{}

	for _, v := range entries {
		groupMap[v.ID] = v
		appGroup := filterAppsByGroup(apps, v.ID)

		for _, h := range v.Hosts {
			hostLookup[h] = v.ID
		}

		switch v.State {
		case "run":
			log.Infof("%s is in state run idle for %d seconds", v.ID, currentTS-v.LatestUsage)

			if scaleMap, isDifferent := scaleMapDiffers(appGroup, v); isDifferent {
				log.Infof("%s update scale map", v.ID)
				hostManager.UpdateScaleMap(v.ID, scaleMap)
			}

			if (v.LatestUsage + v.IdleDuration) < currentTS {
				s.Hibernate(v.ID, hostManager, m)
			}
		case "hibernate":
			log.Infof("%s is in state hibernate since %d seconds", v.ID, currentTS-v.LatestUsage)
			if (v.LatestUsage + v.IdleDuration) > currentTS {
				log.Infof("%s wakeup", v.ID)
				s.Wakeup(v.ID, hostManager, m)
			}
		}
	}

	s.Lock()
	defer s.Unlock()
	s.HostLookup = hostLookup
	s.GroupMap = groupMap
}
