package web

// Config specific to web package
type Config struct {
	Listen       string
	Location     string
	QueueSize    int
	WorkersCount int
	MyLeader     string
}
