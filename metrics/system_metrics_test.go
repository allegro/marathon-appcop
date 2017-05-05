package metrics

import (
	"testing"

	"runtime"

	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
)

func TestMetricShouldBeGauge(t *testing.T) {
	t.Parallel()

	// expect
	assert.Implements(t, (*metrics.Gauge)(nil), baseGauge{})
}

func TestGaugeReturnsValueFromGivenFunction(t *testing.T) {
	t.Parallel()

	// given
	var counter int64
	bg := baseGauge{value: func(_ runtime.MemStats) int64 {
		counter++
		return counter
	}}

	//when
	bg.Value()
	bg.Value()
	bg.Value()

	//then
	assert.Equal(t, (int64)(3), counter)
}

func TestMetricsRegistered(t *testing.T) {
	t.Parallel()

	//when
	collectSystemMetrics()

	//then
	assert.NotNil(t, metrics.Get(systemMetric(allocGauge)))
	assert.NotNil(t, metrics.Get(systemMetric(heapObjectsGauge)))
	assert.NotNil(t, metrics.Get(systemMetric(totalPauseGauge)))
	assert.NotNil(t, metrics.Get(systemMetric(lastPauseGauge)))
}
