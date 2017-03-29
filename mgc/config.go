package mgc

import "time"

// Config specific to mgc module
type Config struct {
	MaxSuspendTime time.Duration
	Interval       time.Duration
	AppCopOnly     bool
	Enabled        bool
}
