package marathon

import (
	"errors"
	"net/url"
)

// MStub is a stub for marathon functionality
type MStub struct {
	Apps   []*App
	Groups []*Group
	// AppGetFail - Set to true and this Stub get method will return errors
	AppsGetFail bool
	// AppDelFail - Set to true and this Stub methods will return error
	AppDelFail bool
	// AppDelHalfFail - Set to true and this Stub methods will return error at each second call
	AppDelHalfFail   bool
	GroupDelFail     bool
	AppScaleDownFail bool
	FailCounter      *FailCounter
	ScaleCounter     *ScaleCounter
}

// FailCounter is structure to hold state between failures
type FailCounter struct {
	Counter int
}

// ScaleCounter is counting scaling operations
type ScaleCounter struct {
	Counter int
}

// AppsGet get stubbed apps
func (m MStub) AppsGet() ([]*App, error) {
	if m.AppsGetFail {
		return nil, errors.New("Unable to get applications from marathon")
	}
	return m.Apps, nil
}

// AppGet get stubbed app
func (m MStub) AppGet(appID AppID) (*App, error) {
	return &App{ID: appID}, nil
}

// GroupsGet get stubbed groups
func (m MStub) GroupsGet() ([]*Group, error) {
	return m.Groups, nil
}

// TasksGet get stubed Tasks
func (m MStub) TasksGet(appID AppID) ([]*Task, error) {
	return []*Task{
		{AppID: appID},
	}, nil
}

// AuthGet get stubbed auth
func (m MStub) AuthGet() *url.Userinfo {
	return &url.Userinfo{}
}

// LocationGet get stubbed location
func (m MStub) LocationGet() string {
	return ""
}

// LeaderGet get stubbed leader
func (m MStub) LeaderGet() (string, error) {
	return "", nil
}

// AppScaleDown by one instance
func (m MStub) AppScaleDown(app *App) error {
	if m.AppScaleDownFail {
		return errors.New("Unable to scale down")
	}
	m.ScaleCounter.Counter = 1
	return nil
}

// AppDelete application by provided AppID
func (m MStub) AppDelete(appID AppID) error {
	if m.AppDelFail {
		return errors.New("Unable to delete app")
	}
	if m.AppDelHalfFail {
		if m.FailCounter.Counter%2 == 0 {
			m.FailCounter.Counter++
			return errors.New("Unable to delete app")
		}
		m.FailCounter.Counter++
	}
	return nil
}

// GroupDelete by provided GroupID
func (m MStub) GroupDelete(groupID GroupID) error {
	if m.GroupDelFail {
		return errors.New("Unable to delete group")
	}
	return nil
}

// GetEmptyLeafGroups returns groups from marathon which are leafs in group tree
func (m MStub) GetEmptyLeafGroups() ([]*Group, error) {
	return []*Group{}, nil
}
