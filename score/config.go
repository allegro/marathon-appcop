package score

import "time"

// Config contains specific configuration to score module
type Config struct {
	DryRun           bool
	ScaleDownScore   int
	UpdateInterval   time.Duration
	ResetInterval    time.Duration
	EvaluateInterval time.Duration
	ScaleLimit       int
}
