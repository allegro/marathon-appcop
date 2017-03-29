package marathon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/sethgrid/pester"

	log "github.com/Sirupsen/logrus"
	"github.com/allegro/marathon-appcop/metrics"
)

// Marathoner interfacing marathon
type Marathoner interface {
	AppGet(AppID) (*App, error)
	AppsGet() ([]*App, error)
	TasksGet(AppID) ([]*Task, error)
	AuthGet() *url.Userinfo
	LocationGet() string
	LeaderGet() (string, error)
	AppScaleDown(*App) error
	AppDelete(AppID) error
	GroupsGet() ([]*Group, error)
	GroupDelete(GroupID) error
}

// Marathon reciever
type Marathon struct {
	Location string
	Protocol string
	Auth     *url.Userinfo
	client   *pester.Client
}

// ScaleData marathon scale json representation
type ScaleData struct {
	Instances int               `json:"instances"`
	Labels    map[string]string `json:"labels"`
}

// ScaleResponse represents marathon response from scaling request
type ScaleResponse struct {
	Version      string `json:"version"`
	DeploymentID string `json:"deploymentId"`
}

// DeleteResponse represents marathon response from scaling request
type DeleteResponse struct {
	Version      string `json:"version"`
	DeploymentID string `json:"deploymentId"`
}

// LeaderResponse represents marathon response from /v2/leader request
type LeaderResponse struct {
	Leader string `json:"leader"`
}

type urlParams map[string]string

// New marathon instance
func New(config Config) (*Marathon, error) {
	var auth *url.Userinfo
	if len(config.Username) == 0 && len(config.Password) == 0 {
		auth = nil
	} else {
		auth = url.UserPassword(config.Username, config.Password)
	}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	pClient := pester.New()
	pClient.Concurrency = 3
	pClient.MaxRetries = 5
	pClient.Backoff = pester.ExponentialBackoff
	pClient.KeepLog = true
	pClient.Transport = transport

	return &Marathon{
		Location: config.Location,
		Protocol: config.Protocol,
		Auth:     auth,
		client:   pClient,
	}, nil
}

// AppGet get marathons application from v2/apps/<AppID>
func (m Marathon) AppGet(appID AppID) (*App, error) {
	log.WithField("Location", m.Location).Debugf("Asking Marathon for %s", appID)

	body, err := m.get(m.urlWithQuery(fmt.Sprintf("/v2/apps/%s", appID), urlParams{"embed": "apps.tasks"}))
	if err != nil {
		return nil, err
	}

	return ParseApp(body)
}

// AppsGet get marathons application from v2/apps/<AppID>
func (m Marathon) AppsGet() ([]*App, error) {
	log.Debug("Asking Marathon for list of applications")

	body, err := m.get(m.url("/v2/apps/"))
	if err != nil {
		return nil, err
	}

	return ParseApps(body)
}

// TasksGet lists marathon tasks for specified AppID
func (m Marathon) TasksGet(appID AppID) ([]*Task, error) {
	log.WithFields(log.Fields{
		"Location": m.Location,
		"Id":       appID,
	}).Debug("asking Marathon for tasks")

	trimmedAppID := strings.Trim(appID.String(), "/")
	body, err := m.get(m.url(fmt.Sprintf("/v2/apps/%s/tasks", trimmedAppID)))
	if err != nil {
		return nil, err
	}

	return ParseTasks(body)
}

func close(r *http.Response) {
	err := r.Body.Close()
	if err != nil {
		log.WithError(err).Error("Can't close response")
	}
}

func (m Marathon) get(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	request.Header.Add("Accept", "application/json")

	log.WithFields(log.Fields{
		"Uri":      request.URL.RequestURI(),
		"Location": m.Location,
		"Protocol": m.Protocol,
	}).Debug("Sending GET request to Marathon")

	var response *http.Response
	metrics.Time("marathon.get", func() {
		response, err = m.client.Do(request)
	})
	if err != nil {
		metrics.Mark("marathon.get.error")
		m.logHTTPError(response, err)
		return nil, err
	}
	defer close(response)
	if response.StatusCode != 200 {
		metrics.Mark("marathon.get.error")
		metrics.Mark(fmt.Sprintf("marathon.get.error.%d", response.StatusCode))
		err = fmt.Errorf("Expected 200 but got %d for %s", response.StatusCode, response.Request.URL.Path)
		m.logHTTPError(response, err)
		return nil, err
	}

	return ioutil.ReadAll(response.Body)
}

func (m Marathon) update(url string, d []byte) ([]byte, error) {
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(d))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	request.Header.Add("Accept", "application/json")

	log.WithFields(log.Fields{
		"Uri":      request.URL.RequestURI(),
		"Location": m.Location,
		"Protocol": m.Protocol,
	}).Debug("Sending PUT request to marathon")

	var response *http.Response
	metrics.Time("marathon.put", func() {
		response, err = m.client.Do(request)
	})
	if err != nil {
		log.Warn("Updating application failed.")
		metrics.Mark("marathon.put.error")
		m.logHTTPError(response, err)
		return nil, err
	}
	defer close(response)

	if response.StatusCode != 200 {
		metrics.Mark("marathon.put.error")
		metrics.Mark(fmt.Sprintf("marathon.put.error.%d", response.StatusCode))
		err = fmt.Errorf("Expected 200 but got %d for %s", response.StatusCode, response.Request.URL.Path)
		m.logHTTPError(response, err)
		return nil, err
	}

	return ioutil.ReadAll(response.Body)
}

func (m Marathon) delete(url string) ([]byte, error) {
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	request.Header.Add("Accept", "application/json")

	log.WithFields(log.Fields{
		"Uri":      request.URL.RequestURI(),
		"Location": m.Location,
		"Protocol": m.Protocol,
	}).Debug("Sending DELETE request to marathon")

	var response *http.Response
	metrics.Time("marathon.delete", func() {
		response, err = m.client.Do(request)
	})
	if err != nil {
		log.Warn("Deleting application failed.")
		metrics.Mark("marathon.delete.error")
		m.logHTTPError(response, err)
		return nil, err
	}
	defer close(response)

	if response.StatusCode != 200 {
		metrics.Mark("marathon.delete.error")
		metrics.Mark(fmt.Sprintf("marathon.delete.error.%d", response.StatusCode))
		err = fmt.Errorf("Expected 200 but got %d for %s", response.StatusCode, response.Request.URL.Path)
		m.logHTTPError(response, err)
		return nil, err
	}

	return ioutil.ReadAll(response.Body)
}

// AppScaleDown scales down app by provided AppID
func (m Marathon) AppScaleDown(app *App) error {
	var instances int

	log.WithFields(log.Fields{
		"AppID": app.ID,
	}).Debug("Scaling Down application because of score.")

	if app.Instances >= 1 {
		instances = app.Instances - 1
	} else {
		return errors.New("Unable to scale down, zero instance")
	}

	// rewrite labels and add new one
	labels := app.Labels
	if instances == 0 {
		labels["appcop"] = "suspend"
	} else {
		labels["appcop"] = "scaleDown"
	}
	log.WithFields(log.Fields{
		"AppID": app,
		"label": labels["appcop"],
	}).Info("Altering labels.")

	scaleData := &ScaleData{Instances: instances, Labels: labels}
	u, err := json.Marshal(scaleData)
	if err != nil {
		return err
	}

	trimmedAppID := strings.Trim(app.ID.String(), "/")
	url := m.urlWithQuery(fmt.Sprintf("/v2/apps/%s", trimmedAppID),
		urlParams{"force": "true"})

	log.WithFields(log.Fields{
		"url": url,
	}).Debug("Application url.")

	body, err := m.update(url, u)
	if err != nil {
		return err
	}

	scaleResponse := &ScaleResponse{}
	return json.Unmarshal(body, scaleResponse)
}

// AppDelete scales down app by provided AppID
func (m Marathon) AppDelete(app AppID) error {

	log.WithFields(log.Fields{
		"AppID": app,
	}).Info("Deleting application.")

	trimmedAppID := strings.Trim(app.String(), "/")
	url := m.url(fmt.Sprintf("/v2/apps/%s", trimmedAppID))

	log.WithFields(log.Fields{
		"url": url,
	}).Debug("Application url.")

	body, err := m.delete(url)
	if err != nil {
		return err
	}

	deleteResponse := &DeleteResponse{}
	return json.Unmarshal(body, deleteResponse)
}

func (m Marathon) logHTTPError(resp *http.Response, err error) {
	statusCode := "???"
	if resp != nil {
		statusCode = fmt.Sprintf("%d", resp.StatusCode)
	}

	log.WithFields(log.Fields{
		"Location":   m.Location,
		"Protocol":   m.Protocol,
		"statusCode": statusCode,
	}).Error(err)
}

func (m Marathon) url(path string) string {
	return m.urlWithQuery(path, nil)
}

func (m Marathon) urlWithQuery(path string, params urlParams) string {
	marathon := url.URL{
		Scheme: m.Protocol,
		User:   m.Auth,
		Host:   m.Location,
		Path:   path,
	}
	query := marathon.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	marathon.RawQuery = query.Encode()
	return marathon.String()
}

// GroupsGet get marathons application from v2/apps/<AppID>
func (m Marathon) GroupsGet() ([]*Group, error) {
	log.Debug("Asking Marathon for list of groups")

	body, err := m.get(m.url("/v2/groups/"))
	if err != nil {
		return nil, err
	}

	return ParseGroups(body)
}

// GroupDelete scales down app by provided AppID
func (m Marathon) GroupDelete(group GroupID) error {

	log.WithFields(log.Fields{
		"GroupID": group,
	}).Info("Deleting group.")

	trimmedGroupID := strings.Trim(group.String(), "/")
	url := m.url(fmt.Sprintf("/v2/groups/%s", trimmedGroupID))

	log.WithFields(log.Fields{
		"url": url,
	}).Debug("Group url.")

	body, err := m.delete(url)
	if err != nil {
		return err
	}

	deleteResponse := &DeleteResponse{}
	return json.Unmarshal(body, deleteResponse)
}

// AuthGet string from marathon configured instance
func (m Marathon) AuthGet() *url.Userinfo {
	return m.Auth
}

// LocationGet from marathon configured instance
func (m Marathon) LocationGet() string {
	return m.Location
}

// LeaderGet from marathon cluster
func (m Marathon) LeaderGet() (string, error) {
	log.WithField("Location", m.Location).Debug("Asking Marathon for leader")
	body, err := m.get(m.url("/v2/leader"))
	if err != nil {
		return "", err
	}
	leaderResponse := &LeaderResponse{}
	err = json.Unmarshal(body, leaderResponse)

	return leaderResponse.Leader, err
}
