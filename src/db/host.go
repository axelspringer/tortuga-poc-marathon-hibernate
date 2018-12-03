package db

import (
	"fmt"
)

// HostState enum type
type HostState int

const (
	// HostStateRun state
	HostStateRun HostState = iota
	// HostStateHibernate state
	HostStateHibernate
	// HostStateWakeUp state
	HostStateWakeUp
	// HostStateDozeOff state
	HostStateDozeOff
)

// String representation
func (hs HostState) String() string {
	return [...]string{"run", "hibernate", "wakeup", "dozeoff"}[hs]
}

// Parse string to HostState
func (hs *HostState) Parse(s string) error {
	switch s {
	case "run":
		*hs = HostStateRun
		return nil
	case "hibernate":
		*hs = HostStateHibernate
		return nil
	case "wakeup":
		*hs = HostStateWakeUp
		return nil
	case "dozeoff":
		*hs = HostStateDozeOff
		return nil
	}

	return fmt.Errorf("HostState parse: unknown type %s", s)
}

// HostEntry model
type HostEntry struct {
	Host           string `json:"host"`
	LatestTrigger  int    `json:"latestTrigger"`
	State          string `json:"state"`
	HibernateGroup string `json:"group"`
}

// HostEntryManager interface
type HostEntryManager interface {
	GetAllEntries() []*HostEntry
	GetByHost(host string) *HostEntry
	GetByState(state HostState) []*HostEntry
}
