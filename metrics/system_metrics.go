package metrics

import (
	"runtime"

	"github.com/rcrowley/go-metrics"
)

const allocGauge = "runtime.mem.bytes_allocated_and_not_yet_freed"
const heapObjectsGauge = "runtime.mem.total_number_of_allocated_objects"
const totalPauseGauge = "runtime.mem.pause_total_ns"
const lastPauseGauge = "runtime.mem.last_pause"

func collectSystemMetrics() {
	err := metrics.Register(
		systemMetric(allocGauge), baseGauge{value: func(memStats runtime.MemStats) int64 { return int64(memStats.Alloc) }})
	if err != nil {
		return
	}
	err = metrics.Register(
		systemMetric(heapObjectsGauge), baseGauge{value: func(memStats runtime.MemStats) int64 { return int64(memStats.HeapObjects) }})
	if err != nil {
		return
	}
	err = metrics.Register(
		systemMetric(totalPauseGauge), baseGauge{value: func(memStats runtime.MemStats) int64 { return int64(memStats.PauseTotalNs) }})
	if err != nil {
		return
	}
	err = metrics.Register(
		systemMetric(lastPauseGauge), baseGauge{value: func(memStats runtime.MemStats) int64 { return int64(memStats.PauseNs[(memStats.NumGC+255)%256]) }})
	if err != nil {
		return
	}
}

type baseGauge struct {
	value func(runtime.MemStats) int64
}

func (g baseGauge) Value() int64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return g.value(memStats)
}

func (g baseGauge) Snapshot() metrics.Gauge { return metrics.GaugeSnapshot(g.Value()) }

func (baseGauge) Update(int64) {}
