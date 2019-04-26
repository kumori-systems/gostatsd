package sepagent

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
	"./sepastats"
	"./udpclient"

	"github.com/atlassian/gostatsd"

	log "github.com/sirupsen/logrus"
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
	config *Config,
	updclient *udpclient.UDPClient,
	sepastats *sepastats.SEPAStats
}

// Name returns the name of the backend.
func (Client) Name() string {
	return backendName
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
	updclient *udpclient.UDPClient, err := udpclient.New(config.Host, config.Port)
	if err != nil {
		return
	}
	sepastats *sepastats.SEPAStats, err := sepastats.New(config.SepRegExpStr)
	if err != nil {
		return
	}
	client = &Client{
		config: config,
		udpclient: udpclient,
		sepastats: sepastats
	}
	return
}

// SendMetricsAsync prints the metrics in a MetricsMap to the sepagent, preparing payload synchronously but doing the send asynchronously.
func (client Client) SendMetricsAsync(ctx context.Context, metrics *gostatsd.MetricMap, cb gostatsd.SendCallback) {
	go func() {
		cb([]error{})
	}()
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


/*
// SendMetricsAsync prints the metrics in a MetricsMap to the sepagent, preparing payload synchronously but doing the send asynchronously.
func (client Client) SendMetricsAsync(ctx context.Context, metrics *gostatsd.MetricMap, cb gostatsd.SendCallback) {
	buf := client.preparePayload(metrics)
	go func() {
		cb([]error{writePayload(buf)})
	}()
}

func writePayload(buf *bytes.Buffer) (retErr error) {
	writer := log.StandardLogger().Writer()
	defer func() {
		if err := writer.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	_, err := writer.Write(buf.Bytes())
	return err
}

func preparePayload(metrics *gostatsd.MetricMap) *bytes.Buffer {
	myCounter := 0
	myTimer := 0
	myGauge := 0
	SEPKEY := "envoy.cluster.external_cluster_"
	buf := new(bytes.Buffer)
	now := time.Now().Unix()
	metrics.Counters.Each(func(key, tagsKey string, counter gostatsd.Counter) {
		if strings.Contains(key, SEPKEY) {
			myCounter++
		}
	})
	metrics.Timers.Each(func(key, tagsKey string, timer gostatsd.Timer) {
		if strings.Contains(key, SEPKEY) {
			myTimer++
		}
	})
	metrics.Gauges.Each(func(key, tagsKey string, gauge gostatsd.Gauge) {
		if strings.Contains(key, SEPKEY) {
			myGauge++
		}
	})
	fmt.Fprintf(buf, "{ myCounter: %d, myTimer: %d, myGauge: %d, time: %d }\n", myCounter, myTimer, myGauge, now)
	return buf
}

// SendEvent prints events to the sepagent.
func (client Client) SendEvent(ctx context.Context, e *gostatsd.Event) (retErr error) {
	writer := log.StandardLogger().Writer()
	defer func() {
		if err := writer.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	_, err := fmt.Fprintf(writer, "event: %+v\n", e)
	return err
}
*/