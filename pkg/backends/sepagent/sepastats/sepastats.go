// Package sepastats is used to convert gostatsd metrics into more convenient
// structures for SepAgent, and manipulate them as serialized JSON objects.
package sepastats

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Item contains stats info in a convenient format for SepAgent
type Item struct {
	InstanceID      string
	Cluster         string
	MetricType      string
	AggregationType string
	Statistics      string
	Value           float64
	UnixTimestamp   int64
}

// SEPAStats is used to convert gostatsd metrics into more convenient
// structures for SepAgent, and manipulate them as serialized JSON objects.
type SEPAStats struct {
	sepRegExp *regexp.Regexp
}

// New creates a new SEPAStats
func New(sepRegExpStr string) (s SEPAStats, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("SEPAStats.New: %s", r.(error).Error())
		}
	}()
	sepRegExp := regexp.MustCompile("_" + sepRegExpStr + "$")
	s = SEPAStats{sepRegExp}
	return
}

// CreateItem constructs an Item object using gostatsd record properties
// For example, something like this in gostatsd metrics ...
//   metricType = "counter"
//   key = "envoy.cluster.external_cluster_sep1.upstream_rq_time"
//   aggregation = "count"
//   value = 100
// ... is converted to Item
//   {
//   	 instanceID = "sep1",
//   	 cluster = "external_cluster_sep1",
//   	 metricType = "counter",  (counter, gauge or timers)
//   	 aggregationType = "count",   (count, per_second, ....)
//     statistic = "upstram_rq_time"
//   	 value = 100
//   	 unixTimestamp = 1556100401
//   }
func (s SEPAStats) CreateItem(metricType string, key string, aggregation string, value float64, timestamp int64) (item Item, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sepastats: %s", r.(error).Error())
		}
	}()
	keyParts := strings.Split(key, ".")
	cluster := keyParts[len(keyParts)-2]
	statistic := keyParts[len(keyParts)-1]
	if s.sepRegExp.MatchString(cluster) == false {
		err = fmt.Errorf("sepastats: Sep regular expression not found in key %s", key)
		return
	}
	item = Item{}
	item.Statistics = statistic
	item.InstanceID = s.sepRegExp.FindString(cluster)[1:]
	item.Cluster = cluster
	item.MetricType = metricType
	item.AggregationType = aggregation
	item.Value = float64(value)
	item.UnixTimestamp = timestamp
	return item, nil
}

// JSONSerialize converts an Item in a serialized ([]byte) JSON
func (s SEPAStats) JSONSerialize(item Item) (b []byte, err error) {
	b, err = json.Marshal(item)
	return
}
