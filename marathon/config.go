package marathon

import "time"

// Config contains marathon module specific configuration
type Config struct {
	Location string
	Protocol string
	Username string
	Password string
	// AppIDPrefix is a part of application id preferably present
	// in all applications in marathon, if found it is removed for the sake of
	// making applications paths shorter.
	// By default this string is empty and no prefix is considered.
	AppIDPrefix string
	VerifySsl   bool
	Timeout     time.Duration
}
