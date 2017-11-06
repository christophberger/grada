package grada

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// ## The data aggregator

// Count is a single time series data tuple, consisting of
// a floating-point value N and a timestamp T.
type Count struct {
	N float64
	T time.Time
}

// Metric is a ring buffer of Counts. It collects time series data that a Grafana
// dashboard panel can request at regular intervals.
// Each Metric has a name that Grafana uses for selecting the desired data stream.
// See Dashboard.CreateMetric().
type Metric struct {
	m        sync.Mutex
	list     []Count
	head     int
	unsorted bool // AddWithTime() and AddCount() do not add in a sorted manner.
}

// Add a single value to the Metric buffer, along with the current time stamp.
// When the buffer is full, every new value overwrites the oldest one.
func (g *Metric) Add(n float64) {
	g.m.Lock()
	defer g.m.Unlock()
	g.list[g.head] = Count{n, time.Now()}
	g.head = (g.head + 1) % len(g.list)
}

// AddWithTime adds a single (value, timestamp) tuple to the ring buffer.
func (g *Metric) AddWithTime(n float64, t time.Time) {
	g.AddCount(Count{n, t})
}

// AddCount adds a complete Count object to the metric data.
func (g *Metric) AddCount(c Count) {
	g.m.Lock()
	defer g.m.Unlock()
	g.unsorted = true
	g.list[g.head] = c
	g.head = (g.head + 1) % len(g.list)
}

// sort sorts the list of metrics by timestamp.
// if the list is already sorted, sort() is a no-op.
func (g *Metric) sort() {
	if !g.unsorted {
		return
	}

	// the ring buffer is unsorted.

	// sooner implements the less func for sort.Slice.
	sooner := func(i, j int) bool {
		return g.list[i].T.UnixNano() < g.list[j].T.UnixNano()
	}

	sort.Slice(g.list, sooner)
	g.head = 0
	g.unsorted = false
}

// fetchDatapoints is called by the Web API server.
// It extracts all datapoints from g.list that fall within the time range [from, to],
// with at most maxDataPoints items.
func (g *Metric) fetchDatapoints(from, to time.Time, maxDataPoints int) *[]row {

	g.m.Lock()
	defer g.m.Unlock()
	length := len(g.list)

	g.sort()

	// Stage 1: extract all data points within the given time range.
	pointsInRange := make([]row, 0, length)
	for i := 0; i < length; i++ {
		count := g.list[(i+g.head)%length] // wrap around
		if count.T.After(from) && count.T.Before(to) {
			pointsInRange = append(pointsInRange, row{count.N, count.T.UnixNano() / 1000000}) // need ms
		}
	}

	points := len(pointsInRange)

	if points <= maxDataPoints {
		return &pointsInRange
	}

	// Stage 2: if more data points than requested exist in the time range,
	// thin out the slice evenly
	rows := make([]row, maxDataPoints)
	ratio := float64(len(pointsInRange)) / float64(len(rows))
	for i := range rows {
		rows[i] = pointsInRange[int(float64(i)*ratio)]
	}

	return &rows
}

// metrics is a map of all metric buffers, with the key being the target name.
// Used internally by the HTTP server and the dashboard.
type metrics struct {
	m      sync.Mutex
	metric map[string]*Metric
}

// Get gets the metric with name "target" from the Metrics map. If a metric of that name
// does not exists in the map, Get returns an error.
func (m *metrics) Get(target string) (*Metric, error) {
	m.m.Lock()
	mt, ok := m.metric[target]
	m.m.Unlock()
	if !ok {
		return nil, errors.New("no such metric: " + target)
	}
	return mt, nil
}

// Put adds a Metric to the Metrics map. Adding an already existing metric
// is an error.
func (m *metrics) Put(target string, metric *Metric) error {
	m.m.Lock()
	defer m.m.Unlock()

	_, exists := m.metric[target]
	if exists {
		return errors.New("metric " + target + " already exists")
	}
	m.metric[target] = metric
	return nil
}

// Delete removes a metric from the Metrics map. Deleting a non-existing
// metric is an error.
func (m *metrics) Delete(target string) error {
	m.m.Lock()
	defer m.m.Unlock()
	_, exists := m.metric[target]
	if !exists {
		return errors.New("cannot delete metric: " + target + " does not exist")
	}
	delete(m.metric, target)
	return nil
}

// Create creates a new Metric with the given target name and buffer size
// and adds it to the Metrics map.
// If a metric for target "target" exists already, Create returns an error.
func (m *metrics) Create(target string, size int) (*Metric, error) {
	metric := &Metric{
		list: make([]Count, size, size),
	}
	err := m.Put(target, metric)
	return metric, err
}
