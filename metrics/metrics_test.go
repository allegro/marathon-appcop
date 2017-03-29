package metrics

import (
	"fmt"
	"os"
	"testing"

	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
)

func TestMark(t *testing.T) {
	// given
	err := Init(Config{Target: "stdout", Prefix: ""})

	// expect
	assert.Nil(t, metrics.Get("marker"))

	// when
	Mark("marker")

	// then
	mark, _ := metrics.Get("marker").(metrics.Meter)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), mark.Count())

	// when
	Mark("marker")

	// then
	assert.Equal(t, int64(2), mark.Count())
}

func TestTime(t *testing.T) {
	// given
	err := Init(Config{Target: "stdout", Prefix: ""})

	// expect
	assert.Nil(t, metrics.Get("timer"))

	// when
	Time("timer", func() {})

	// then
	time, _ := metrics.Get("timer").(metrics.Timer)
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

	// expect
	assert.Nil(t, metrics.Get("counter"))

	// when
	UpdateGauge("counter", 2)

	// then
	gauge := metrics.Get("counter").(metrics.Gauge)
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
