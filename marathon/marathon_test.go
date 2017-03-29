package marathon

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarathonTasksWhenMarathonConnectionFailedShouldNotRetry(t *testing.T) {
	t.Parallel()
	// given
	calls := 0
	server, transport := mockServer(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	})
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1
	// when
	tasks, err := m.TasksGet("/app/id")
	//then
	m.client.Concurrency = 1
	m.client.MaxRetries = 1
	assert.Error(t, err)
	assert.Empty(t, tasks)
	assert.Equal(t, 1, calls)
}

func TestMarathonAppGetWhenMarathonConnectionFailedShouldNotRetry(t *testing.T) {
	t.Parallel()
	// given
	calls := 0
	server, transport := mockServer(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	})
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1
	// when
	app, err := m.AppGet("/app/id")
	//then
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Equal(t, 1, calls)
}

func TestMarathonAppsWhenMarathonReturnMalformedJSONResponse(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps/testapp?embed=apps.tasks", `{"apps":}`)
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	// when
	app, err := m.AppGet("/testapp")
	//then
	assert.Nil(t, app)
	assert.Error(t, err)
}

func TestMarathonAppWhenMarathonReturnEmptyApp(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps//test/app?embed=apps.tasks", `{"app": {}}`)
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	// when
	app, err := m.AppGet("/test/app")
	//then
	assert.NoError(t, err)
	assert.NotNil(t, app)
}

func TestMarathonAppWhenMarathonReturnEmptyResponse(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps//test/app?embed=apps.tasks", ``)
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	// when
	app, err := m.AppGet("/test/app")
	//then
	assert.NotNil(t, app)
	assert.Error(t, err)
}

func TestMarathonTasksWhenMarathonReturnEmptyList(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps/test/app/tasks", `
	{"tasks": [{
		"appId": "/test",
		"host": "192.0.2.114",
		"id": "test.47de43bd-1a81-11e5-bdb6-e6cb6734eaf8",
		"ports": [31315],
		"healthCheckResults":[{ "alive":true }]
	}]}`)
	defer server.Close()
	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	// when
	tasks, err := m.TasksGet("/test/app")
	//then
	assert.NoError(t, err)
	assert.NotNil(t, tasks)
}

func TestMarathonTasksWhenMarathonReturnEmptyResponse(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps/test/app/tasks", ``)
	defer server.Close()
	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	// when
	tasks, err := m.TasksGet("/test/app")
	//then
	assert.Nil(t, tasks)
	assert.Error(t, err)
}

func TestMarathonTasksWhenMarathonReturnMalformedJSONResponse(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps/test/app/tasks", ``)
	defer server.Close()
	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	// when
	tasks, err := m.TasksGet("/test/app")
	//then
	assert.Nil(t, tasks)
	assert.Error(t, err)
}

func TestURLWithoutAuth(t *testing.T) {
	t.Parallel()
	// given
	config := Config{Location: "example.com:8080", Protocol: "http"}
	// when
	m, _ := New(config)
	// then
	assert.Equal(t, "http://example.com:8080/v2/apps", m.url("/v2/apps"))
}

func TestURLWithAuth(t *testing.T) {
	t.Parallel()
	// given
	config := Config{Location: "example.com:8080", Protocol: "http", Username: "peter", Password: "parker"}
	// when
	m, _ := New(config)
	// then
	assert.Equal(t, "http://peter:parker@example.com:8080/v2/apps", m.url("/v2/apps"))
}

func TestLeaderGetRetrunsNoLeader(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/leader", `{"leader":}`)
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1

	// when
	leader, err := m.LeaderGet()
	//then
	assert.Equal(t, leader, "")
	assert.Error(t, err)
}

func TestLeaderGetRetrunsCorrectLeader(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/leader", `{"leader": "marathon-leader"}`)
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1

	// when
	leader, err := m.LeaderGet()
	//then
	require.NoError(t, err)
	assert.Equal(t, leader, "marathon-leader")
}

func TestGetLeaderRetrunsMalformedJSON(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/leader", `{"leader": `)
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1

	// when
	leader, err := m.LeaderGet()
	//then
	require.Error(t, err)
	assert.Equal(t, "", leader)
}

func TestMarathonGetLeaderWhenMarathonConnectionFailedShouldNotRetry(t *testing.T) {
	t.Parallel()
	// given
	calls := 0
	server, transport := mockServer(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	})
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1
	// when
	leader, err := m.LeaderGet()
	//then
	assert.Error(t, err)
	assert.Equal(t, leader, "")
	assert.Equal(t, 1, calls)
}

func TestGetLocationReturnsLocation(t *testing.T) {
	t.Parallel()
	// given
	m, _ := New(Config{Location: "marathon:8080"})
	// when
	location := m.LocationGet()
	//then
	assert.Equal(t, "marathon:8080", location)

}

func TestGetLocationReturnsNoLocation(t *testing.T) {
	t.Parallel()
	// given
	m, _ := New(Config{})
	// when
	location := m.LocationGet()
	//then
	assert.Equal(t, "", location)
}

func TestGetAuthReturnsCorrectAuth(t *testing.T) {
	t.Parallel()
	// given
	m, _ := New(Config{Username: "test", Password: "test"})
	// when
	auth := m.AuthGet()
	//then
	assert.Equal(t, "test:test", auth.String())

}

func TestGetAuthReturnsNoAuth(t *testing.T) {
	t.Parallel()
	// given
	m, err := New(Config{})
	// when
	auth := m.AuthGet()
	//then
	assert.NoError(t, err)
	assert.Nil(t, auth)
}

func TestMarathonAppsGetWhenMarathonConnectionFailedShouldNotRetry(t *testing.T) {
	t.Parallel()
	// given
	calls := 0
	server, transport := mockServer(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	})
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1
	// when
	app, err := m.AppsGet()
	//then
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Equal(t, 1, calls)
}

func TestMarathonAppsGetWhenMarathonReturnEmptyApp(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps/", `{"apps": {}}`)
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	// when
	app, err := m.AppsGet()
	//then
	assert.Error(t, err)
	assert.Nil(t, app)
}

func TestMarathonAppsGetWhenReturnsMalformedJSON(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps/", `{"apps": `)
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	// when
	app, err := m.AppsGet()
	//then
	assert.Error(t, err)
	assert.Nil(t, app)
}

func TestMarathonAppScaleDownWhenMarathonConnectionFailedShouldNotRetry(t *testing.T) {
	t.Parallel()
	// given
	calls := 0
	server, transport := mockServer(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	})
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1
	// when
	app := &App{
		ID: "testapp", Instances: 1,
		Labels: make(map[string]string),
	}
	err := m.AppScaleDown(app)
	//then
	assert.Error(t, err)
}

func TestMarathonScaleDownAppsSuccess(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps/testapp0?force=true",
		`{"version": "0", "deploymentId": "a"}`)
	defer server.Close()
	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport

	app := &App{
		ID: "testapp0", Instances: 2,
		Labels: make(map[string]string),
	}

	// when
	err := m.AppScaleDown(app)
	//then
	assert.Nil(t, err)

}

func TestMarathonScaleDownAppsZeroInstances(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps/testapp0?force=true",
		`{"version": "0",
			"deploymentId": "a"}`,
	)
	defer server.Close()
	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport

	app := &App{
		ID: "testapp0", Instances: 0,
		Labels: make(map[string]string),
	}

	// when
	err := m.AppScaleDown(app)
	//then
	assert.Error(t, err)
}

func TestMarathonAppDeleteSuccess(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/apps/testapp",
		`{"version": "0",
			"deploymentId": "a"}`,
	)
	defer server.Close()
	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport

	// when
	err := m.AppDelete("testapp")
	//then
	assert.Nil(t, err)
}

func TestMarathonAppDeleteWhenMarathonReturns500(t *testing.T) {
	t.Parallel()
	// given
	calls := 0
	server, transport := mockServer(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	})
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1
	// when
	err := m.AppDelete("testapp")
	//then
	assert.Error(t, err)
}

func TestMarathonGroupsGetSuccessMarathonReturnsOneGroup(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/groups/",
		`{"groups": [
			{"apps": [], "groups": [], "id": "idgroup0", "version": "1234"}
			]}`,
	)
	defer server.Close()
	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport

	// when
	groups, err := m.GroupsGet()
	//then
	require.NoError(t, err)
	require.NotNil(t, groups)
	assert.Equal(t, groups[0].Version, "1234")
	assert.Equal(t, groups[0].ID.String(), "idgroup0")
}

func TestMarathonGroupsGetWhenMarathonReturns500(t *testing.T) {
	t.Parallel()
	// given
	calls := 0
	server, transport := mockServer(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	})
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1
	// when
	_, err := m.GroupsGet()
	//then
	assert.Error(t, err)
}

func TestMarathonGroupDeleteSuccessOnExampleGroup(t *testing.T) {
	t.Parallel()
	// given
	server, transport := stubServer("/v2/groups/testgroup",
		`{"version": "0",
			"deploymentId": "a"}`,
	)
	defer server.Close()
	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport

	// when
	err := m.GroupDelete("testgroup")
	//then
	require.NoError(t, err)
}

func TestMarathonGroupDeleteWhenMarathonReturns500(t *testing.T) {
	t.Parallel()
	// given
	calls := 0
	server, transport := mockServer(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
	})
	defer server.Close()

	url, _ := url.Parse(server.URL)
	m, _ := New(Config{Location: url.Host, Protocol: "HTTP"})
	m.client.Transport = transport
	m.client.Concurrency = 1
	m.client.MaxRetries = 1
	// when
	err := m.GroupDelete("testgroup")
	//then
	assert.Error(t, err)
}

// http://keighl.com/post/mocking-http-responses-in-golang/
func stubServer(uri string, body string) (*httptest.Server, *http.Transport) {
	return mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() == uri {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, body)
		} else {
			w.WriteHeader(404)
		}
	})
}

func mockServer(handle func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *http.Transport) {
	server := httptest.NewServer(http.HandlerFunc(handle))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	return server, transport
}
