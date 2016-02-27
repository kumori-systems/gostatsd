package stdout

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jtblin/gostatsd/backend"
	"github.com/jtblin/gostatsd/types"

	log "github.com/Sirupsen/logrus"
)

const backendName = "stdout"

func init() {
	backend.RegisterBackend(backendName, func() (backend.MetricSender, error) {
		return NewClient()
	})
}

// Client is an object that is used to send messages to stdout
type Client struct{}

// NewClient constructs a StdoutClient object
func NewClient() (backend.MetricSender, error) {
	return &Client{}, nil
}

// Regular expressions used for bucket name normalization
var regSemiColon = regexp.MustCompile(":")

// normalizeBucketName cleans up a bucket name by replacing or translating invalid characters
func normalizeBucketName(bucket string, tagsKey string) string {
	tags := strings.Split(tagsKey, ",")
	for _, tag := range tags {
		if tag != "" {
			bucket += "." + regSemiColon.ReplaceAllString(tag, "_")
		}
	}
	return bucket
}

// SampleConfig returns the sample config for the stdout backend
func (client *Client) SampleConfig() string {
	return ""
}

// SendMetrics sends the metrics in a MetricsMap to the Graphite server
func (client *Client) SendMetrics(metrics types.MetricMap) error {
	buf := new(bytes.Buffer)
	now := time.Now().Unix()
	types.EachCounter(metrics.Counters, func(key, tagsKey string, counter types.Counter) {
		nk := normalizeBucketName(key, tagsKey)
		fmt.Fprintf(buf, "stats.counter.%s.count %d %d\n", nk, counter.Value, now)
		fmt.Fprintf(buf, "stats.counter.%s.per_second %f %d\n", nk, counter.PerSecond, now)
	})
	types.EachTimer(metrics.Timers, func(key, tagsKey string, timer types.Timer) {
		nk := normalizeBucketName(key, tagsKey)
		fmt.Fprintf(buf, "stats.timers.%s.lower %f %d\n", nk, timer.Min, now)
		fmt.Fprintf(buf, "stats.timers.%s.upper %f %d\n", nk, timer.Max, now)
		fmt.Fprintf(buf, "stats.timers.%s.count %d %d\n", nk, timer.Count, now)
		fmt.Fprintf(buf, "stats.timers.%s.count_ps %f %d\n", nk, timer.PerSecond, now)
		fmt.Fprintf(buf, "stats.timers.%s.mean %f %d\n", nk, timer.Mean, now)
		fmt.Fprintf(buf, "stats.timers.%s.median %f %d\n", nk, timer.Median, now)
		fmt.Fprintf(buf, "stats.timers.%s.sum %f %d\n", nk, timer.Sum, now)
		fmt.Fprintf(buf, "stats.timers.%s.sum %f %d\n", nk, timer.SumSquares, now)
		fmt.Fprintf(buf, "stats.timers.%s.sum_squares %f %d\n", nk, timer.StdDev, now)
		for _, pct := range timer.Percentiles {
			fmt.Fprintf(buf, "stats.timers.%s.%s %f %d\n", nk, pct.String(), pct.Float(), now)
		}
	})
	types.EachGauge(metrics.Gauges, func(key, tagsKey string, gauge types.Gauge) {
		nk := normalizeBucketName(key, tagsKey)
		fmt.Fprintf(buf, "stats.gauge.%s %f %d\n", nk, gauge.Value, now)
	})

	types.EachSet(metrics.Sets, func(key, tagsKey string, set types.Set) {
		nk := normalizeBucketName(key, tagsKey)
		fmt.Fprintf(buf, "stats.set.%s %d %d\n", nk, len(set.Values), now)
	})

	fmt.Fprintf(buf, "statsd.numStats %d %d\n", metrics.NumStats, now)
	fmt.Fprintf(buf, "statsd.processingTime %f %d\n", float64(metrics.ProcessingTime)/float64(time.Millisecond), now)

	_, err := buf.WriteTo(log.StandardLogger().Writer())
	if err != nil {
		return err
	}
	return nil
}

// Name returns the name of the backend
func (client *Client) Name() string {
	return backendName
}
