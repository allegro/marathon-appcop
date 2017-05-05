package metrics

import (
	"fmt"
	"os"
	"testing"

	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var filterOutEmptyStringsTestCases = []struct {
	input, expectedOutput []string
}{
	{
		input:          []string{"a", ""},
		expectedOutput: []string{"a"},
	},
	{
		input:          []string{"a", "b"},
		expectedOutput: []string{"a", "b"},
	},
	{
		input:          []string{"", ""},
		expectedOutput: nil,
	},
}

func TestFilterOutEmptyStrings(t *testing.T) {
	t.Parallel()

	for _, testCase := range filterOutEmptyStringsTestCases {
		actualOutput := FilterOutEmptyStrings(testCase.input)
		assert.Equal(t, testCase.expectedOutput, actualOutput)
	}
}

var systemMetricTestcases = []struct {
	config         Config
	metric         string
	expectedMetric string
}{
	{
		config:         Config{SystemSubPrefix: "system"},
		metric:         "metric",
		expectedMetric: "system.localhost.metric",
	},
	{
		config:         Config{SystemSubPrefix: ""},
		metric:         "metric",
		expectedMetric: "localhost.metric",
	},
	{
		config:         Config{SystemSubPrefix: ""},
		metric:         "",
		expectedMetric: "localhost",
	},
}

func TestSystemMetricTestCases(t *testing.T) {
	hostname = func() (string, error) { return "localhost", nil }
	for _, testCase := range systemMetricTestcases {
		err := Init(testCase.config)
		require.NoError(t, err)

		actualMetric := systemMetric(testCase.metric)
		assert.Equal(t, testCase.expectedMetric, actualMetric)
	}
}

var appMetricTestcases = []struct {
	config         Config
	metric         string
	expectedMetric string
}{
	{
		config:         Config{AppSubPrefix: "applications"},
		metric:         "metric",
		expectedMetric: "applications.metric",
	},
	{
		config:         Config{AppSubPrefix: ""},
		metric:         "metric",
		expectedMetric: "metric",
	},
	{
		config:         Config{AppSubPrefix: ""},
		metric:         "",
		expectedMetric: "",
	},
}

func TestAppMetricTestCases(t *testing.T) {
	hostname = func() (string, error) { return "localhost", nil }
	for _, testCase := range appMetricTestcases {
		err := Init(testCase.config)
		require.NoError(t, err)

		actualMetric := appMetric(testCase.metric)
		assert.Equal(t, testCase.expectedMetric, actualMetric)
	}
}

func TestMark(t *testing.T) {
	// given
	err := Init(Config{Target: "stdout", Prefix: ""})
	systemMarker := systemMetric("marker")

	// expect
	assert.Nil(t, metrics.Get(systemMarker))

	// when
	Mark("marker")

	// then
	mark, _ := metrics.Get(systemMarker).(metrics.Meter)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), mark.Count())

	// when
	Mark("marker")

	// then
	assert.Equal(t, int64(2), mark.Count())
}

func TestMarkApp(t *testing.T) {
	// given
	err := Init(Config{Target: "stdout", Prefix: "prefix", AppSubPrefix: "applications"})
	appMarker := "applications.marker"

	// expect
	assert.Nil(t, metrics.Get(appMarker))

	// when
	MarkApp("marker")

	// then
	mark, _ := metrics.Get(appMarker).(metrics.Meter)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), mark.Count())

	// when
	MarkApp("marker")

	// then
	assert.Equal(t, int64(2), mark.Count())
}

func TestTime(t *testing.T) {
	// given
	err := Init(Config{Target: "stdout", Prefix: ""})
	systemTimer := systemMetric("timer")

	// expect
	assert.Nil(t, metrics.Get(systemTimer))

	// when
	Time("timer", func() {})

	// then
	time, _ := metrics.Get(systemTimer).(metrics.Timer)
	assert.Equal(t, int64(1), time.Count())

	// when
	Time("timer", func() {})

	// then
	assert.Nil(t, err)
	assert.Equal(t, int64(2), time.Count())
}

func TestUpdateGauge(t *testing.T) {
	// given
	err := Init(Config{Target: "stdout", Prefix: ""})
	systemCounter := systemMetric("counter")

	// expect
	assert.Nil(t, metrics.Get(systemCounter))

	// when
	UpdateGauge("counter", 2)

	// then
	gauge := metrics.Get(systemCounter).(metrics.Gauge)
	assert.Equal(t, int64(2), gauge.Value())

	// when
	UpdateGauge("counter", 123)

	// then
	assert.Equal(t, int64(123), gauge.Value())
	assert.Nil(t, err)
}

func TestMetricsInit_ForGraphiteWithNoAddress(t *testing.T) {
	err := Init(Config{Target: "graphite", Addr: ""})
	assert.Error(t, err)
}

func TestMetricsInit_ForGraphiteWithBadAddress(t *testing.T) {
	err := Init(Config{Target: "graphite", Addr: "localhost"})
	assert.Error(t, err)
}

func TestMetricsInit_ForGraphit(t *testing.T) {
	err := Init(Config{Target: "graphite", Addr: "localhost:81"})
	assert.NoError(t, err)
}

func TestMetricsInit_ForUnknownTarget(t *testing.T) {
	err := Init(Config{Target: "unknown"})
	assert.Error(t, err)
}

func TestMetricsInit(t *testing.T) {
	// when
	err := Init(Config{Prefix: "prefix"})

	// then
	assert.Equal(t, "prefix", prefix)
	assert.NoError(t, err)
}

func TestInit_DefaultPrefix(t *testing.T) {
	// given
	hostname = func() (string, error) { return "", fmt.Errorf("Some error") }

	// when
	err := Init(Config{Prefix: "default"})

	// then
	assert.Error(t, err)
}

func TestInit_DefaultPrefix_WithErrors(t *testing.T) {
	// given
	hostname = func() (string, error) { return "myhost", nil }
	os.Args = []string{"./myapp"}

	// when
	err := Init(Config{Prefix: "default"})

	// then
	assert.NoError(t, err)
	assert.Equal(t, "myhost.myapp", prefix)
}
