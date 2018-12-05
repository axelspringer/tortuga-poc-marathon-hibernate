package marathon

import (
	"net/url"
	"time"

	marathon "github.com/gambol99/go-marathon"
	log "github.com/sirupsen/logrus"
)

// ApplicationFilterFunc filters application lists
type ApplicationFilterFunc func(a *marathon.Application) *marathon.Application

// Manager model
type Manager struct {
	connector *Connector
}

// NewManager creates a new manager
func NewManager(connector *Connector) *Manager {
	m := &Manager{
		connector: connector,
	}

	return m
}

// ListAppsWithHibernationSupport list applications with hibernation support
func (m *Manager) ListAppsWithHibernationSupport() []marathon.Application {
	apps, err := m.connector.Client.Applications(url.Values{})

	if err != nil {
		log.Warnf("Could not find any applications. %s", err)
		return []marathon.Application{}
	}

	filtered := filterApplications(apps, func(a *marathon.Application) *marathon.Application {
		if a.Labels == nil {
			return nil
		}

		if v, ok := (*a.Labels)["hiberthon.enable"]; !ok || v != "true" {
			return nil
		}

		return a
	})

	return filtered
}

// Deploy applications
func (m *Manager) Deploy(apps []marathon.Application) {
	for _, app := range apps {
		// dirty hack
		if app.Uris != nil && len(*app.Uris) > 0 {
			app.Fetch = nil
		}

		log.Infof("Deploying %s", app.ID)
		_, err := m.connector.Client.ScaleApplicationInstances(app.ID, *app.Instances, true)
		if err != nil {
			log.Errorf("Error %s", err)
			continue
		}
	}

	for _, app := range apps {
		m.connector.Client.WaitOnApplication(app.ID, 30*time.Second)
	}

	log.Infof("Deployment done\n")
}

// ListAppsByLableGroup list applications with group association
func (m *Manager) ListAppsByLableGroup(groupID string) []marathon.Application {
	apps, err := m.connector.Client.Applications(url.Values{})

	if err != nil {
		log.Warnf("Could not find any applications. %s", err)
		return []marathon.Application{}
	}

	filtered := filterApplications(apps, func(a *marathon.Application) *marathon.Application {
		if a.Labels == nil {
			return nil
		}

		if v, ok := (*a.Labels)["hiberthon.enable"]; !ok || v != "true" {
			return nil
		}

		if v, ok := (*a.Labels)["hiberthon.group"]; !ok || v != groupID {
			return nil
		}

		return a
	})

	return filtered
}

func filterApplications(a *marathon.Applications, filter ApplicationFilterFunc) []marathon.Application {
	resList := []marathon.Application{}
	for _, app := range a.Apps {
		if fa := filter(&app); fa != nil {
			resList = append(resList, app)
		}
	}
	return resList
}
