package metrics

import "time"

// Config specific to metrics package
type Config struct {
	Target   string
	Prefix   string
	Interval time.Duration
	Addr     string
}
