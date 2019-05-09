// Package sepastats is used to convert gostatsd metrics into more convenient
// structures for SepAgent, and manipulate them as serialized JSON objects.
package sepastats

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// Item contains stats info in a convenient format for SepAgent
type Item struct {
	InstanceID      string  `json:"instanceId"`
	Cluster         string  `json:"cluster"`
	MetricType      string  `json:"metricType"`
	AggregationType string  `json:"aggregationType"`
	Statistics      string  `json:"statistics"`
	Value           float64 `json:"value"`
	UnixTimestamp   int64   `json:"unixTimestamp"`
}

// Just for json encoding.
type itemEx struct {
	Sep     string `json:"sep"`
	Metrics []Item `json:"metrics"`
}

// SEPAStats is used to convert gostatsd metrics into more convenient
// structures for SepAgent, and get them as serialized JSON objects.
type SEPAStats struct {
	sepRegExp     *regexp.Regexp
	clusterRegExp *regexp.Regexp
	seps          map[string][]Item
}

// -----------------------------------------------------------------------------
// PUBLIC METHODS
// -----------------------------------------------------------------------------

// New creates a new SEPAStats
func New(sepRegExpStr string, clusterRegExpStr string) (s *SEPAStats, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("SEPAStats.New: %s", r.(error).Error())
		}
	}()
	sepRegExp := regexp.MustCompile(sepRegExpStr)
	seps := make(map[string][]Item)
	clusterRegExp := regexp.MustCompile(clusterRegExpStr)
	s = &SEPAStats{sepRegExp, clusterRegExp, seps}
	return
}

// AddItem constructs an Item object using gostatsd record properties,
// and adds it to the seps map.
// For example, something like this in gostatsd metrics ...
//   metricType = "counter"
//   key = "envoy.cluster.external_cluster_eslap.cloud_sep_1.upstream_rq_time"
//   aggregation = "count"
//   value = 100
// ... is converted to Item
//   {
//   	 instanceID = "sep1",
//   	 cluster = "external_cluster_eslap.cloud_sep_1",
//   	 metricType = "counter",  (counter, gauge or timers)
//   	 aggregationType = "count",   (count, per_second, ....)
//     statistic = "upstram_rq_time"
//   	 value = 100
//   	 unixTimestamp = 1556100401
//   }
func (s *SEPAStats) AddItem(metricType string, key string, aggregation string, value float64, timestamp int64) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sepastats: %s", r.(error).Error())
		}
	}()
	cluster := s.clusterRegExp.FindString(key)
	sepInstance := s.sepRegExp.FindString(cluster)
	if cluster == "" || sepInstance == "" {
		err = fmt.Errorf("sepastats: sep or cluster not found in key %s", key)
		return
	}
	clusterIndex := s.clusterRegExp.FindStringIndex(key)
	statistic := key[clusterIndex[1]+1:]
	item := Item{}
	item.Statistics = statistic
	item.InstanceID = sepInstance
	item.Cluster = cluster
	item.MetricType = metricType
	item.AggregationType = aggregation
	item.Value = float64(value)
	item.UnixTimestamp = timestamp
	s.addToMap(item)
	return
}

// IsSep checks if the provided key corresponds to a sep metric
// Example of key: "envoy.cluster.external_cluster_eslap.cloud_sep_1.upstream_rq_time"
func (s *SEPAStats) IsSep(key string) (rc bool) {
	defer func() {
		if r := recover(); r != nil {
			rc = false
		}
	}()
	clusterCheck := (s.clusterRegExp.MatchString(key) == true)
	sepCheck := (s.sepRegExp.MatchString(key) == true)
	rc = (clusterCheck && sepCheck)
	return
}

// GetItems returns the map of seps metrics
func (s *SEPAStats) GetItems() map[string][]Item {
	return s.seps
}

// ClearItems removes all stored data
func (s *SEPAStats) ClearItems() {
	s.seps = make(map[string][]Item)
}

// GetSerializedSep converts a sep entry in a serialized ([]byte) JSON, like
// this:
//
func (s *SEPAStats) GetSerializedSep(sep string) (b []byte, err error) {
	if s.seps[sep] == nil {
		err = fmt.Errorf("sepastats: Sep %s not found", sep)
		return
	}
	aux := &itemEx{
		Sep:     sep,
		Metrics: s.seps[sep],
	}
	b, err = json.Marshal(aux)
	return
}

// -----------------------------------------------------------------------------
// PRIVATE METHODS
// -----------------------------------------------------------------------------

// addItem adds an item to the seps map
func (s *SEPAStats) addToMap(item Item) {
	if s.seps[item.InstanceID] == nil {
		s.seps[item.InstanceID] = make([]Item, 0, 10)
	}
	s.seps[item.InstanceID] = append(s.seps[item.InstanceID], item)
}
