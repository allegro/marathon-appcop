package marathon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAppsRecievesMalformedJSONBlob(t *testing.T) {
	t.Parallel()
	// given
	var appsJSON = []byte(`{"apps": {}`)
	//	// when
	apps, err := ParseApps(appsJSON)
	//	//then
	require.Error(t, err)
	assert.Nil(t, apps)
}

func TestParseTaskRecievesCorrectJSONBlob(t *testing.T) {
	t.Parallel()
	// given
	var taskJSON = []byte(`{"taskStatus": "healthy",
      "appId": "appid",
      "host": "fqdn" }
    `)
	//	// when
	task, err := ParseTask(taskJSON)
	//	//then
	expected := &Task{ID: "",
		TaskStatus:         "healthy",
		AppID:              "appid",
		Host:               "fqdn",
		Ports:              []int(nil),
		HealthCheckResults: []HealthCheckResult(nil),
	}
	require.NoError(t, err)
	assert.Equal(t, expected, task)
}

func TestParseTaskRecievesMalformedJSONBlob(t *testing.T) {
	t.Parallel()
	// given
	var taskJSON = []byte(`{"taskStatus": {`)
	//	// when
	_, err := ParseTask(taskJSON)
	//	//then
	require.Error(t, err)
}

func TestParseTasksRecievesCorrectJSONBlob(t *testing.T) {
	t.Parallel()
	// given
	var tasksJSON = []byte(`{"tasks": [
      { "taskStatus": "healthy",
        "appId": "appid",
        "host": "fqdn" },
      { "taskStatus": "unlealthy",
        "appId": "appid",
        "host": "fqdn" }
    ]
  }`)
	//	// when
	task, err := ParseTasks(tasksJSON)
	//	//then
	expected := []*Task{
		{ID: "",
			TaskStatus:         "healthy",
			AppID:              "appid",
			Host:               "fqdn",
			Ports:              []int(nil),
			HealthCheckResults: []HealthCheckResult(nil)},
		{ID: "",
			TaskStatus:         "unlealthy",
			AppID:              "appid",
			Host:               "fqdn",
			Ports:              []int(nil),
			HealthCheckResults: []HealthCheckResult(nil)},
	}
	require.NoError(t, err)
	assert.Equal(t, expected, task)
}

func TestParseTasksRecievesMalformedJSONBlob(t *testing.T) {
	t.Parallel()
	// given
	var taskJSON = []byte(`"tasks": [ {}]tus": {`)
	//	// when
	_, err := ParseTasks(taskJSON)
	//	//then
	require.Error(t, err)
}
