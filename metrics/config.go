package metrics

import "time"

// Config specific to metrics package
type Config struct {
	Target string
	//Prefix is the begining of metric, it is prepended
	// in each and every published metric.
	Prefix   string
	Interval time.Duration
	Addr     string
	Instance string
	// SystemSubPrefix it is part of a metric that is appended to the
	// main Prefix, representing appcop internal metrics
	// essential to appcop admins, e.g runtime metrics, event processing time,
	// event queue size etc.
	SystemSubPrefix string
	// AppSubPrefix it is part of a metric that is appended to the
	// main Prefix, representing applications specific metric, e.g task_running,
	// task_staging, task_failed.
	AppSubPrefix string
}
