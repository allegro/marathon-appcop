package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/allegro/marathon-appcop/marathon"
	"github.com/allegro/marathon-appcop/metrics"
	"github.com/allegro/marathon-appcop/mgc"
	"github.com/allegro/marathon-appcop/score"
	"github.com/allegro/marathon-appcop/web"
	flag "github.com/ogier/pflag"
)

// Config specific config
type Config struct {
	Web      web.Config
	Marathon marathon.Config
	Score    score.Config
	MGC      mgc.Config
	Metrics  metrics.Config
	Log      struct {
		Level  string
		Format string
		File   string
	}
	configFile string
}

var config = &Config{}

// NewConfig config instance
func NewConfig() (*Config, error) {

	if !flag.Parsed() {
		config.parseFlags()
	}
	flag.Parse()
	err := config.loadConfigFromFile()

	if err != nil {
		return nil, err
	}

	err = config.setLogOutput()
	if err != nil {
		return nil, err
	}

	config.setLogFormat()
	err = config.setLogLevel()
	if err != nil {
		return nil, err
	}

	return config, err
}

func (config *Config) parseFlags() {
	// Web
	flag.StringVar(&config.Web.Listen, "listen", ":4444", "Port to listen on, at this point only for health checking")
	flag.StringVar(&config.Web.Location, "event-stream", "http://example.com:8080/v2/events", "Get events from this stream")
	flag.IntVar(&config.Web.QueueSize, "events-queue-size", 1000, "Size of events queue")
	flag.IntVar(&config.Web.WorkersCount, "workers-pool-size", 10, "Number of concurrent workers processing events")
	flag.StringVar(&config.Web.MyLeader, "my-leader", "example.com:8080", "My leader, when marathon /v2/leader endpoint return the same string as this one, make subscription to event stream")

	// Marathon
	flag.StringVar(&config.Marathon.Location,
		"marathon-location", "example.com:8080", "Marathon URL")
	flag.StringVar(&config.Marathon.Protocol,
		"marathon-protocol", "http", "Marathon protocol (http or https)")
	flag.StringVar(&config.Marathon.Username,
		"marathon-username", "marathon", "Marathon username for basic auth")
	flag.StringVar(&config.Marathon.Password,
		"marathon-password", "marathon", "Marathon password for basic auth")
	flag.BoolVar(&config.Marathon.VerifySsl,
		"marathon-ssl-verify", true, "Verify certificates when connecting via SSL")
	flag.DurationVar(&config.Marathon.Timeout,
		"marathon-timeout", 30*time.Second,
		"Time limit for requests made by the Marathon HTTP client. A Timeout of zero means no timeout")
	flag.StringVar(&config.Marathon.AppIDPrefix, "appid-prefix", "",
		"Prefix common to all fully qualified application ID's. Remove this preffix from applications id's (reffer to README to get an idea when this id is removed)")

	// Score
	flag.BoolVar(&config.Score.DryRun,
		"dry-run", false,
		"Perform a trial run with no changes made to marathon.")
	flag.IntVar(&config.Score.ScaleDownScore,
		"scale-down-score", 200,
		"Score for application to scale it one instance down.")
	flag.IntVar(&config.Score.ScaleLimit,
		"scale-limit", 2,
		"How many application scale down actions to commit in one EvaluateInterval.")
	flag.DurationVar(&config.Score.UpdateInterval,
		"update-interval", 2*time.Second,
		"Interval of updating app scores.")
	flag.DurationVar(&config.Score.ResetInterval,
		"reset-interval", 60*time.Minute,
		"Interval when apps are scored, after interval passes scores are reset.")
	flag.DurationVar(&config.Score.EvaluateInterval,
		"evaluate-interval", 2*time.Minute,
		"Interval when apps are scored, after interval passes scores are reset.")

	// Marathon GC
	flag.BoolVar(&config.MGC.Enabled,
		"mgc-enabled", true,
		"Enable garbage collecting of marathon, old suspended applications will be deleted.")
	flag.DurationVar(&config.MGC.MaxSuspendTime,
		"mgc-max-suspend-time", 7*24*time.Hour,
		"How long application should be suspended before deleting it.")
	flag.DurationVar(&config.MGC.Interval,
		"mgc-interval", 8*time.Hour,
		"Marathon GC interval.")
	flag.BoolVar(&config.MGC.AppCopOnly,
		"mgc-appcop-only", true,
		"Delete only applications suspended by appcop.")

	// Metrics
	flag.StringVar(&config.Metrics.Target, "metrics-target", "stdout",
		"Metrics destination stdout or graphite (empty string disables metrics)")
	flag.StringVar(&config.Metrics.Prefix, "metrics-prefix", "default",
		"Metrics prefix (default is resolved to <hostname>.<app_name>")
	flag.StringVar(&config.Metrics.SystemSubPrefix, "metrics-system-sub-prefix", "appcop-internal",
		"System specific metrics. Append to metric-prefix")
	flag.StringVar(&config.Metrics.AppSubPrefix, "metrics-app-sub-prefix", "applications",
		"Applications specific metrics. Appended to metric-prefix")
	flag.DurationVar(&config.Metrics.Interval, "metrics-interval", 30*time.Second,
		"Metrics reporting interval")
	flag.StringVar(&config.Metrics.Addr, "metrics-location", "",
		"Graphite URL (used when metrics-target is set to graphite)")
	flag.StringVar(&config.Metrics.Addr, "metrics-instance", "",
		"Part of Graphite metric, used to distinguish between AppCop instances internal metrics.")

	// Log
	flag.StringVar(&config.Log.Level, "log-level", "info",
		"Log level: panic, fatal, error, warn, info, or debug")
	flag.StringVar(&config.Log.Format, "log-format", "text",
		"Log format: JSON, text")
	flag.StringVar(&config.Log.File, "log-file", "",
		"Save logs to file (e.g.: `/var/log/appcop.log`). If empty logs are published to STDERR")

	// General
	flag.StringVar(&config.configFile, "config-file", "",
		"Path to a JSON file to read configuration from. Note: Will override options set earlier on the command line")
}

func (config *Config) loadConfigFromFile() error {
	if config.configFile == "" {
		return nil
	}
	jsonBlob, err := ioutil.ReadFile(config.configFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBlob, config)
}

func (config *Config) setLogLevel() error {
	level, err := log.ParseLevel(config.Log.Level)
	if err != nil {
		log.WithError(err).WithField("Level", config.Log.Level).Error("Bad level")
		return err
	}
	log.SetLevel(level)
	return nil
}

func (config *Config) setLogOutput() error {
	path := config.Log.File

	if len(path) == 0 {
		log.SetOutput(os.Stderr)
		return nil
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.WithError(err).Errorf("error opening file: %s", path)
		return err
	}

	log.SetOutput(f)
	return nil
}

func (config *Config) setLogFormat() {
	format := strings.ToUpper(config.Log.Format)
	if format == "JSON" {
		log.SetFormatter(&log.JSONFormatter{})
	} else if format == "TEXT" {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.WithField("Format", format).Error("Unknown log format")
	}
}
