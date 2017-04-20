package marathon

import (
	"encoding/json"
	"fmt"
	"strings"
)

const ApplicationImmunityLabel = "APP_IMMUNITY"

// AppWrapper json returned from marathon with app definition
type AppWrapper struct {
	App App `json:"app"`
}

// AppsResponse json returned from marathon with apps definitions
type AppsResponse struct {
	Apps []*App `json:"apps"`
}

// App represents application returned in marathon json
type App struct {
	Labels      map[string]string `json:"labels"`
	ID          AppID             `json:"id"`
	Tasks       []Task            `json:"tasks"`
	Instances   int               `json:"instances"`
	VersionInfo VersionInfo       `json:"versionInfo"`
}

// HasImmunity check if application behavior is tolerated without consequence
func (app App) HasImmunity() bool {
	if val, ok := app.Labels[ApplicationImmunityLabel]; ok && val == "true" {
		return true
	}
	return false
}

func (app *App) penalize() error {

	if app.Instances >= 1 {
		app.Instances--
	} else {
		return fmt.Errorf("Unable to scale down, zero instance")
	}

	if app.Instances == 0 {
		app.Labels["appcop"] = "suspend"
	} else {
		app.Labels["appcop"] = "scaleDown"
	}

	return nil

}

// VersionInfo represents json field of  this name, inside marathon app
// definition
type VersionInfo struct {
	LastScalingAt      string `json:"lastScalingAt"`
	LastConfigChangeAt string `json:"lastConfigChangeAt"`
}

// AppID Marathon Application Id (aka PathId)
// Usually in the form of /rootGroup/subGroup/subSubGroup/name
// allowed characters: lowercase letters, digits, hyphens, slash
type AppID string

// String stringer for app
func (id AppID) String() string {
	return string(id)
}

// ParseApps json
func ParseApps(jsonBlob []byte) ([]*App, error) {
	apps := &AppsResponse{}
	err := json.Unmarshal(jsonBlob, apps)

	return apps.Apps, err
}

// ParseApp json
func ParseApp(jsonBlob []byte) (*App, error) {
	wrapper := &AppWrapper{}
	err := json.Unmarshal(jsonBlob, wrapper)

	return &wrapper.App, err
}

// Task definition returned in marathon event
type Task struct {
	ID                 TaskID
	TaskStatus         string              `json:"taskStatus"`
	AppID              AppID               `json:"appId"`
	Host               string              `json:"host"`
	Ports              []int               `json:"ports"`
	HealthCheckResults []HealthCheckResult `json:"healthCheckResults"`
}

// TaskID from marathon
// Usually in the form of AppID.uuid with '/' replaced with '_'
type TaskID string

func (id TaskID) String() string {
	return string(id)
}

// AppID contains string defining application in marathon
func (id TaskID) AppID() AppID {
	index := strings.LastIndex(id.String(), ".")
	return AppID("/" + strings.Replace(id.String()[0:index], "_", "/", -1))
}

// HealthCheckResult returned from marathon api
type HealthCheckResult struct {
	Alive bool `json:"alive"`
}

// TasksResponse response to TasksGet call
type TasksResponse struct {
	Tasks []*Task `json:"tasks"`
}

// ParseTasks try to convert raw Tasks data to json
func ParseTasks(jsonBlob []byte) ([]*Task, error) {
	tasks := &TasksResponse{}
	err := json.Unmarshal(jsonBlob, tasks)

	return tasks.Tasks, err
}

// ParseTask try to convert raw Task data to json
func ParseTask(event []byte) (*Task, error) {
	task := &Task{}
	err := json.Unmarshal(event, task)
	return task, err
}
