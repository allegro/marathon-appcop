package marathon

import (
	"fmt"
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

var penalizeTestCases = []struct {
	app         *App
	expectedApp *App
	expectedErr error
}{
	{
		app: &App{ID: "testApp0", Instances: 1, Labels: map[string]string{}},
		expectedApp: &App{ID: "testApp0",
			Instances: 0,
			Labels:    map[string]string{"appcop": "suspend"}},
		expectedErr: nil,
	},
	{
		app: &App{ID: "testApp1",
			Instances: 2,
			Labels:    map[string]string{},
		},
		expectedApp: &App{ID: "testApp1",
			Instances: 1,
			Labels:    map[string]string{"appcop": "scaleDown"}},
		expectedErr: nil,
	},
	{
		app: &App{ID: "testApp2",
			Instances: 2,
			Labels:    map[string]string{},
		},
		expectedApp: &App{ID: "testApp2",
			Instances: 1,
			Labels:    map[string]string{"appcop": "scaleDown"}},
		expectedErr: nil,
	},
	{
		app: &App{ID: "testApp3",
			Instances: 2,
			Labels:    map[string]string{"APPLABEL": "true"}},
		expectedApp: &App{ID: "testApp3",
			Instances: 1,
			Labels:    map[string]string{"appcop": "scaleDown", "APPLABEL": "true"}},
		expectedErr: nil,
	},
	{
		app: &App{ID: "testApp4",
			Instances: 0,
			Labels:    map[string]string{"APPLABEL": "true"}},
		expectedApp: &App{ID: "testApp4",
			Instances: 0,
			Labels:    map[string]string{"APPLABEL": "true"}},
		expectedErr: fmt.Errorf("Unable to scale down, zero instance"),
	},
}

func TestPenalizeTestCases(t *testing.T) {
	for _, testCase := range penalizeTestCases {
		err := testCase.app.penalize()
		require.Equal(t, testCase.expectedErr, err)
		assert.Equal(t, testCase.app, testCase.expectedApp)
	}
}

func TestIsImmuneShouldReturnTrueWhenImmunityLabelIsSetTrue(t *testing.T) {
	// given
	app := &App{
		Labels: map[string]string{ApplicationImmunityLabel: "true"},
	}
	// when
	actualExcused := app.HasImmunity()
	expectedExcused := true
	// then
	assert.Equal(t, expectedExcused, actualExcused)
}
