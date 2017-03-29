package marathon

import "time"

// Config contains marathon module specific configuration
type Config struct {
	Location  string
	Protocol  string
	Username  string
	Password  string
	VerifySsl bool
	Timeout   time.Duration
}
