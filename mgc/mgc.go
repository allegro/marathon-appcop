package mgc

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/allegro/marathon-appcop/marathon"
	"github.com/allegro/marathon-appcop/metrics"
)

// MarathonGC is Marathon Garbage Collector receiever, mainly holds applications registry
// and marathon client
type MarathonGC struct {
	config      Config
	marathon    marathon.Marathoner
	apps        []*marathon.App
	lastRefresh time.Time
}

// New instantiates MarathonGC reciever
func New(config Config, marathon marathon.Marathoner) (*MarathonGC, error) {

	return &MarathonGC{
		config:      config,
		marathon:    marathon,
		apps:        nil,
		lastRefresh: time.Time{},
	}, nil
}

// StartMarathonGCJob is highest control element of MarathonGC module,
// which starts job goroutine for periodic:
// - collection of suspended apps,
// - collection of empty groups.
func (mgc *MarathonGC) StartMarathonGCJob() {
	if !mgc.config.Enabled {
		log.Info("Marathon Garbage Collection enabled")
		return
	}
	log.WithFields(log.Fields{
		"Interval": mgc.config.Interval,
	}).Info("Marathon GC job started")

	go func() {
		var err error
		ticker := time.NewTicker(mgc.config.Interval)
		for range ticker.C {
			metrics.Time("mgc.refresh", func() { err = mgc.refresh() })
			if err != nil {
				metrics.Mark("mgc.refresh.error")
				continue
			}
			mgc.gcSuspended()
			mgc.gcEmptyGroups()
		}
	}()
}

// gcSuspended commits garbage collection for suspended apps
func (mgc *MarathonGC) gcSuspended() {
	log.Info("Staring GC on suspended apps")
	apps := mgc.getOldSuspended()
	if len(apps) == 0 {
		log.Info("No suspended apps to gc")
		return
	}

	var deletedCount int
	metrics.Time("mgc.delete.suspended", func() {
		deletedCount = mgc.deleteSuspended(apps)
	})
	if deletedCount == 0 {
		metrics.UpdateGauge("mgc.delete.suspended.count", int64(deletedCount))
		log.Info("Nothing GC'ed for long suspend")
	}

}

// gcEmptyGroups is starting GC jobs on groups
// It is evaluating time of last group update
func (mgc *MarathonGC) gcEmptyGroups() {
	log.Info("Staring GC on empty groups")
	groups, err := mgc.marathon.GroupsGet()
	if err != nil {
		log.WithError(err).Error("Ending GCEmptyGroups")
		return
	}

	for _, group := range groups {
		t, err := toMarathonDate(group.Version)
		if err != nil {
			log.WithError(err).Error("Unable to parse date")
			continue
		}
		if group.IsEmpty() && (t.elapsed() > mgc.config.MaxSuspendTime) {
			metrics.Time("mgc.groups.delete", func() {
				err = mgc.groupDelete(group.ID)
			})
			if err != nil {
				metrics.Mark("mgc.groups.delete.error")
				continue
			}
		}
	}
}

func (mgc *MarathonGC) groupDelete(groupID marathon.GroupID) error {
	log.Infof("Deleting group %s", groupID)
	return mgc.marathon.GroupDelete(groupID)
}

func (mgc *MarathonGC) refresh() error {
	log.WithFields(log.Fields{
		"LastUpdate": mgc.lastRefresh,
	}).Info("Refreshing local app registry")

	// get apps
	apps, err := mgc.marathon.AppsGet()
	if err != nil {
		log.WithFields(log.Fields{
			"LastUpdate": mgc.lastRefresh,
		}).Error("Refresh fail")

		return err
	}
	mgc.apps = apps
	mgc.lastRefresh = time.Now()

	return nil
}

func (mgc *MarathonGC) shouldBeCollected(app *marathon.App) bool {
	if app.Instances > 0 {
		return false
	}
	scaleDate := app.VersionInfo.LastScalingAt
	t, err := toMarathonDate(scaleDate)

	if err != nil {
		log.WithError(err).Error("Unable to parse provided date")
		return false
	}
	return t.elapsed() > mgc.config.MaxSuspendTime
}

func (mgc *MarathonGC) getOldSuspended() []*marathon.App {
	var ret []*marathon.App

	for _, app := range mgc.apps {
		if mgc.shouldBeCollected(app) && (!mgc.config.AppCopOnly || appCopped(app)) {
			ret = append(ret, app)
		}
	}
	return ret
}

// deleteSuspended returns number (int) of successfully deleted applications
func (mgc *MarathonGC) deleteSuspended(apps []*marathon.App) int {

	n := 0
	var err error
	for _, app := range apps {
		err = mgc.marathon.AppDelete(app.ID)
		if err != nil {
			log.WithError(err).Errorf("Error while deleting suspended app: %s", app.ID)
			continue
		}
		n++
	}
	return n
}

func appCopped(app *marathon.App) bool {
	_, ok := app.Labels["appcop"]

	return ok

}
