package sepagent

import (
	"context"
	"fmt"
	"time"

	"github.com/atlassian/gostatsd"
	"github.com/atlassian/gostatsd/pkg/backends/sepagent/sepastats"
	"github.com/atlassian/gostatsd/pkg/backends/sepagent/udpclient"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// BackendName is the name of this backend.
const (
	BackendName = "sepagent"
)

// Config holds configuration for the SepAgent backend
// Example:
//   Host = "localhost"
//	 Port = 5555
//   SepRegExpStr = "sep([0-9]+)"
//   ClusterRegExpStr = "external_cluster_sep([0-9]+)"
type Config struct {
	Host             string
	Port             int
	SepRegExpStr     string
	ClusterRegExpStr string
}

// Client is an object that is used to send messages to sepagent.
type Client struct {
	config    *Config
	udpclient *udpclient.UDPClient
	sepastats *sepastats.SEPAStats
}

// Just for log
func init() {
	fmt.Println("Initializing backends/sepagent module")
}

// NewClientFromViper constructs a sepagent backend.
func NewClientFromViper(v *viper.Viper) (gostatsd.Backend, error) {
	c := getSubViper(v, "sepagent")
	return NewClient(&Config{
		Host:             c.GetString("host"),
		Port:             c.GetInt("port"),
		SepRegExpStr:     c.GetString("sep_regexp"),
		ClusterRegExpStr: c.GetString("cluster_regexp"),
	})
}

// NewClient constructs a sepagent backend.
func NewClient(config *Config) (client *Client, err error) {
	log.Infof("[%s] %s %d %s %s", BackendName, config.Host, config.Port, config.SepRegExpStr, config.ClusterRegExpStr)
	udpclient, err := udpclient.New(config.Host, config.Port)
	if err != nil {
		return
	}
	sepastats, err := sepastats.New(config.SepRegExpStr, config.ClusterRegExpStr)
	if err != nil {
		return
	}
	client = &Client{
		config:    config,
		udpclient: udpclient,
		sepastats: sepastats,
	}
	return
}

// Name returns the name of the backend.
func (Client) Name() string {
	return BackendName
}

// SendMetricsAsync flushes the metrics to the backend, preparing payload
// synchronously but doing the send asynchronously.
// Must not read/write MetricMap asynchronously.
// NOTE: in this implementation "sepagent backend", the send is done
// syncrhonously (is just UDP)
func (client Client) SendMetricsAsync(ctx context.Context, metrics *gostatsd.MetricMap, cb gostatsd.SendCallback) {
	log.Infof("[%s] SendMetricsAsync", BackendName)
	now := time.Now().Unix()
	errs := make([]error, 0, 10)
	s := client.sepastats
	u := client.udpclient
	s.ClearItems()

	metrics.Counters.Each(func(key string, tagsKey string, counter gostatsd.Counter) {
		if s.IsSep(key) {
			if err := s.AddItem("counter", key, "count", float64(counter.Value), now); err != nil {
				errs = append(errs, err)
			}
			if err := s.AddItem("counter", key, "per_second", counter.PerSecond, now); err != nil {
				errs = append(errs, err)
			}
		}
	})

	metrics.Timers.Each(func(key string, tagsKey string, timer gostatsd.Timer) {
		if s.IsSep(key) {
			if err := s.AddItem("timers", key, "lower", timer.Min, now); err != nil {
				errs = append(errs, err)
			}
			if err := s.AddItem("timers", key, "upper", timer.Max, now); err != nil {
				errs = append(errs, err)
			}
			if err := s.AddItem("timers", key, "count", float64(timer.Count), now); err != nil {
				errs = append(errs, err)
			}
			if err := s.AddItem("timers", key, "count_ps", timer.PerSecond, now); err != nil {
				errs = append(errs, err)
			}
			if err := s.AddItem("timers", key, "mean", timer.Mean, now); err != nil {
				errs = append(errs, err)
			}
			if err := s.AddItem("timers", key, "median", timer.Median, now); err != nil {
				errs = append(errs, err)
			}
			if err := s.AddItem("timers", key, "std", timer.StdDev, now); err != nil {
				errs = append(errs, err)
			}
			if err := s.AddItem("timers", key, "sum", timer.Sum, now); err != nil {
				errs = append(errs, err)
			}
			if err := s.AddItem("timers", key, "sum_squares", timer.SumSquares, now); err != nil {
				errs = append(errs, err)
			}
			for _, pct := range timer.Percentiles {
				if err := s.AddItem("timers", key, pct.Str, pct.Float, now); err != nil {
					errs = append(errs, err)
				}
			}
		}
	})

	metrics.Gauges.Each(func(key string, tagsKey string, gauge gostatsd.Gauge) {
		if s.IsSep(key) {
			if err := s.AddItem("gauge", key, "value", gauge.Value, now); err != nil {
				errs = append(errs, err)
			}
		}
	})

	metrics.Sets.Each(func(key string, tagsKey string, set gostatsd.Set) {
		if s.IsSep(key) {
			if err := s.AddItem("set", key, "len", float64(len(set.Values)), now); err != nil {
				errs = append(errs, err)
			}
		}
	})

	items := s.GetItems()
	log.Infof("[%s] SendMetricsAsync - sending %d items", BackendName, len(items))
	for sep := range items {
		jsonBytes, err := s.GetSerializedSep(sep)
		if err != nil {
			errs = append(errs, err)
			break
		}
		err = u.Send(jsonBytes)
		if err != nil {
			errs = append(errs, err)
			break
		}
	}

	if len(errs) > 0 {
		log.Warningf("[%s] %s", BackendName, errs)
	}

	go func() {
		cb(errs)
	}()
}

// SendEvent sends event to the SepAgent.
// Do nothing in this backend, because:
// 1) Metrics have been sent in SendMetricsAsync method (as in the backend "stdout")
// 2) I don't see this method running at any time. ¿¿??
func (client Client) SendEvent(ctx context.Context, e *gostatsd.Event) (err error) {
	return
}

// Tooling ...

func getSubViper(v *viper.Viper, key string) *viper.Viper {
	n := v.Sub(key)
	if n == nil {
		n = viper.New()
	}
	return n
}
