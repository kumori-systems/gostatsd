package sepagent

import (
	"context"
	"time"

	"./sepastats"
	"./udpclient"

	"github.com/atlassian/gostatsd"

	"github.com/spf13/viper"
)

const (
	backendName = "sepagent"
)

// Config holds configuration for the SepAgent backend
// Example:
//   Host = "localhost"
//	 Port = 5555
//   SepRegExpStr = "sep([0-9]+)"
type Config struct {
	Host         *string
	Port         *int
	SepRegExpStr *string
}

// Client is an object that is used to send messages to sepagent.
type Client struct {
	config    *Config
	updclient *udpclient.UDPClient
	sepastats *sepastats.SEPAStats
}

// NewClientFromViper constructs a sepagent backend.
func NewClientFromViper(v *viper.Viper) (gostatsd.Backend, error) {
	c := getSubViper(v, "sepagent")
	return NewClient(&Config{
		Host:         addr(c.GetString("host")),
		Port:         addrI(c.GetInt("port")),
		SepRegExpStr: addr(c.GetString("sep_regexp")),
	})
}

// NewClient constructs a sepagent backend.
func NewClient(config *Config) (client *Client, err error) {
	updclient, err := udpclient.New(config.Host, config.Port)
	if err != nil {
		return
	}
	sepastats, err := sepastats.New(config.SepRegExpStr)
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
	return backendName
}

// SendMetricsAsync flushes the metrics to the backend, preparing payload
// synchronously but doing the send asynchronously.
// Must not read/write MetricMap asynchronously.
func (client Client) SendMetricsAsync(ctx context.Context, metrics *gostatsd.MetricMap, cb gostatsd.SendCallback) {
	now := time.Now().Unix()
	errs := make([]error, 0, counter)
	c.sepastats.Clear()

	metrics.Counters.Each(func(key string, tagsKey string, counter gostatsd.Counter) {
		if c.sepastats.IsSep(key) {
			if err := c.sepastats.AddItem("counter", key, "count", counter.Value, now); err != nil {
				errs = append(errs, err)
			}
			if err := c.sepastats.AddItem("counter", key, "per_second", counter.PerSecond, now); err != nil {
				errs = append(errs, err)
			}
		}
	})

	metrics.Timers.Each(func(key string, tagsKey string, timer gostatsd.Timers) {
		if c.sepastats.IsSep(key) {
			if err := c.sepastats.AddItem("timers", key, "lower", timer.Min, now); err != nil {
				errs = append(errs, err)
			}
			if err := c.sepastats.AddItem("timers", key, "upper", timer.Max, now); err != nil {
				errs = append(errs, err)
			}
			if err := c.sepastats.AddItem("timers", key, "count", timer.Count, now); err != nil {
				errs = append(errs, err)
			}
			if err := c.sepastats.AddItem("timers", key, "count_ps", timer.PerSecond, now); err != nil {
				errs = append(errs, err)
			}
			if err := c.sepastats.AddItem("timers", key, "mean", timer.Mean, now); err != nil {
				errs = append(errs, err)
			}
			if err := c.sepastats.AddItem("timers", key, "median", timer.Median, now); err != nil {
				errs = append(errs, err)
			}
			if err := c.sepastats.AddItem("timers", key, "std", timer.StdDev, now); err != nil {
				errs = append(errs, err)
			}
			if err := c.sepastats.AddItem("timers", key, "sum", timer.Sum, now); err != nil {
				errs = append(errs, err)
			}
			if err := c.sepastats.AddItem("timers", key, "sum_squares", timer.SumSquares, now); err != nil {
				errs = append(errs, err)
			}
			for _, pct := range timer.Percentiles {
				if err := c.sepastats.AddItem("timers", key, pct.Str, pct.Float.SumSquares, now); err != nil {
					errs = append(errs, err)
				}
			}
		}
	})

	metrics.Gauges.Each(func(key string, tagsKey string, gauge gostatsd.Gauge) {
		if c.sepastats.IsSep(key) {
			if err := c.sepastats.AddItem("gauge", key, "value", gauge.Value, now); err != nil {
				errs = append(errs, err)
			}
		}
	})

	metrics.Sets.Each(func(key string, tagsKey string, set gostatsd.Set) {
		if c.sepastats.IsSep(key) {
			if err := c.sepastats.AddItem("set", key, "len", len(set.Value), now); err != nil {
				errs = append(errs, err)
			}
		}
	})

	go func() {
		cb(errs)
	}()
}

// SendEvent sends event to the SepAgent.
func (client Client) SendEvent(context.Context, *Event) error {
	items := client.sepastats.GetItems()
	for sep := range items {
		if jsonBytes, err = client.sepastats.GetSerializedSep(sep); err != nil {
			break
		}
		if err = u.Send(jsonBytes); err != nil {
			break
		}
	}
	return err
}

// Tooling ...

func addr(s string) *string {
	return &s
}

func addrI(i int) *int {
	return &i
}

func getSubViper(v *viper.Viper, key string) *viper.Viper {
	n := v.Sub(key)
	if n == nil {
		n = viper.New()
	}
	return n
}
