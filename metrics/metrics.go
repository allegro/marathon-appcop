package metrics

//All credits to https://github.com/eBay/fabio/tree/master/metrics
import (
	"errors"
	"fmt"
	logger "log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cyberdelia/go-metrics-graphite"
	"github.com/rcrowley/go-metrics"
)

const (
	PathSeparator   = "/"
	MetricSeparator = "."
)

var (
	prefix          string
	instance        string
	systemSubPrefix string
	appSubPrefix    string
)

func FilterOutEmptyStrings(data []string) []string {
	var parts []string
	for _, part := range data {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func systemMetric(name string) string {
	parts := FilterOutEmptyStrings([]string{systemSubPrefix, instance, name})
	return strings.Join(parts, MetricSeparator)
}

func appMetric(name string) string {
	parts := FilterOutEmptyStrings([]string{appSubPrefix, name})
	return strings.Join(parts, MetricSeparator)
}

// Mark or register Meter on graphite
func Mark(name string) {
	meter := metrics.GetOrRegisterMeter(
		systemMetric(name),
		metrics.DefaultRegistry,
	)
	meter.Mark(1)
}

// Mark or register Meter on graphite
func MarkApp(name string) {
	meter := metrics.GetOrRegisterMeter(
		appMetric(name),
		metrics.DefaultRegistry,
	)
	meter.Mark(1)
}

// Time execution of function
func Time(name string, function func()) {
	timer := metrics.GetOrRegisterTimer(
		systemMetric(name),
		metrics.DefaultRegistry,
	)
	timer.Time(function)
}

// UpdateGauge for provided metric
func UpdateGauge(name string, value int64) {
	gauge := metrics.GetOrRegisterGauge(
		systemMetric(name),
		metrics.DefaultRegistry,
	)
	gauge.Update(value)
}

// Init Metrics
func Init(cfg Config) error {
	prefix = cfg.Prefix
	if prefix == "default" {
		pfx, err := defaultPrefix()
		if err != nil {
			return err
		}
		prefix = pfx
	}

	instance = cfg.Instance
	if instance == "" {
		ins, err := hostname()
		if err != nil {
			ins = "localhost"
		}
		instance = ins
	}

	systemSubPrefix = cfg.SystemSubPrefix
	appSubPrefix = cfg.AppSubPrefix

	collectSystemMetrics()

	switch cfg.Target {
	case "stdout":
		log.Info("Sending metrics to stdout")
		return initStdout(cfg.Interval)
	case "graphite":
		if cfg.Addr == "" {
			return errors.New("metrics: graphite addr missing")
		}

		log.Infof("Sending metrics to Graphite on %s as %q", cfg.Addr, prefix)
		return initGraphite(cfg.Addr, cfg.Interval)
	case "":
		log.Infof("Metrics disabled")
		return nil
	default:
		return fmt.Errorf("Invalid metrics target %s", cfg.Target)
	}
}

func clean(s string) string {
	if s == "" {
		return "_"
	}
	s = strings.Replace(s, ".", "_", -1)
	s = strings.Replace(s, ":", "_", -1)
	return strings.ToLower(s)
}

// stubbed out for testing
var hostname = os.Hostname

func defaultPrefix() (string, error) {
	host, err := hostname()
	if err != nil {
		log.WithError(err).Error("Problem with detecting prefix")
		return "", err
	}
	exe := filepath.Base(os.Args[0])
	return clean(host) + "." + clean(exe), nil
}

func initStdout(interval time.Duration) error {
	logger := logger.New(os.Stderr, "localhost: ", logger.Lmicroseconds)
	go metrics.Log(metrics.DefaultRegistry, interval, logger)
	return nil
}

func initGraphite(addr string, interval time.Duration) error {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return fmt.Errorf("metrics: cannot connect to Graphite: %s", err)
	}

	go graphite.Graphite(metrics.DefaultRegistry, interval, prefix, a)
	return nil
}
