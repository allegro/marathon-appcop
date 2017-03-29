package config

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/allegro/marathon-appcop/marathon"
	"github.com/allegro/marathon-appcop/metrics"
	"github.com/allegro/marathon-appcop/score"
	"github.com/allegro/marathon-appcop/web"
	"github.com/stretchr/testify/assert"
)

func TestConfig_NewReturnsErrorWhenFileNotExist(t *testing.T) {
	clear()

	// given
	os.Args = []string{"./appcop", "--config-file=unknown.json"}

	// when
	_, err := NewConfig()

	// then
	assert.Error(t, err)
}

func TestConfig_NewReturnsErrorWhenFileIsNotJson(t *testing.T) {
	clear()

	// given
	os.Args = []string{"./appcop", "--config-file=config.go"}

	// when
	_, err := NewConfig()

	// then
	assert.Error(t, err)
}

func TestConfig_ShouldReturnErrorForBadLogLevel(t *testing.T) {
	clear()

	// given
	os.Args = []string{"./appcop", "--log-level=bad"}

	// when
	_, err := NewConfig()

	// then
	assert.Error(t, err)
}

func TestConfig_ShouldParseFlags(t *testing.T) {
	clear()

	// given
	os.Args = []string{"./appcop", "--log-level=debug", "--marathon-location=test.host:8080", "--log-format=json", "--my-leader=marathon.dev:8080"}

	// when
	actual, err := NewConfig()

	// then
	assert.NoError(t, err)
	assert.Equal(t, "debug", actual.Log.Level)
	assert.Equal(t, "json", actual.Log.Format)
	assert.Equal(t, "test.host:8080", actual.Marathon.Location)
}

func TestConfig_ShouldUseTextFormatterWhenFormatterIsUnknown(t *testing.T) {
	clear()

	// given
	os.Args = []string{"./appcop", "--log-level=debug", "--log-format=unknown", "--workers-pool-size=10", "--my-leader=marathon.dev:8080"}

	// when
	_, err := NewConfig()

	// then
	assert.NoError(t, err)
}

func TestConfig_ShouldBeMergedWithFileDefaultsAndFlags(t *testing.T) {
	clear()
	expected := &Config{
		Web: web.Config{
			Listen:       ":4444",
			QueueSize:    0,
			WorkersCount: 10,
		},
		Marathon: marathon.Config{
			Location:  "example.com:8080",
			Protocol:  "http",
			Username:  "",
			Password:  "",
			VerifySsl: true,
		},
		Score: score.Config{
			ScaleDownScore:   200,
			UpdateInterval:   2 * time.Second,
			ResetInterval:    24 * time.Hour,
			EvaluateInterval: 20 * time.Second,
			ScaleLimit:       0,
		},
		Metrics: metrics.Config{Target: "stdout",
			Prefix:   "default",
			Interval: 30 * time.Second,
			Addr:     ""},
		Log: struct{ Level, Format, File string }{
			Level:  "info",
			Format: "text",
			File:   "",
		},
		configFile: "testdata/config.json",
	}

	os.Args = []string{"./appcop", "--log-level=debug", "--config-file=testdata/config.json", "--marathon-location=example.com:8080", "--workers-pool-size=10", "--my-leader=marathon.dev:8080"}
	actual, err := NewConfig()

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

// http://stackoverflow.com/a/29169727/1387612
func clear() {
	p := reflect.ValueOf(config).Elem()
	p.Set(reflect.Zero(p.Type()))
}
