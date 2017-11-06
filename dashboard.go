package grada

import (
	"time"
)

type Dashboard struct {
	srv *server
}

// GetDashboard initializes and/or returns the only existing dashboard.
// This also starts the HTTP server that responds to queries from Grafana.
// Default port is 3001. Overwrite this port by setting the environment
// variable GRADA_PORT to the desired port number.
func GetDashboard() *Dashboard {
	d := &Dashboard{}
	d.srv = startServer()
	return d
}

// CreateMetric creates a new metric for the given target, time range, and
// data update interval, and stores this metric in the server.
//
// timeRange is the maximum time range the Grafana dashboard will ask for.
// This depends on the user setting for the dashboard.
//
// interval is the (average) interval in which the data points get delivered.
//
// The quotient of timeRange and interval determines the size of the ring buffer
// that holds the most recent data points.
// Typically, the timeRange of a dashboard request should be much larger than
// the interval for the incoming data.
//
// Creating a metric for an existing target is an error. To replace a metric
// (which is rarely needed), call DeleteMetric first.
func (d *Dashboard) CreateMetric(target string, timeRange, interval time.Duration) (*Metric, error) {
	return d.CreateMetricWithBufSize(target, d.bufSizeFor(timeRange, interval))
}

// CreateMetricWithBufSize creates a new metric for the given target and with the
// given buffer size, and stores this metric in the server.
//
// Buffer size should be chosen so that the buffer can hold enough items for a given
// time range that Grafana asks for and the given rate of data point updates.
//
// Example: If the dashboards's time range is 5 minutes and the incoming data arrives every
// second, the buffer should hold 300 item (5*60*1) at least.
//
// Creating a metric for an existing target is an error. To replace a metric
// (which is rarely needed), call DeleteMetric first.
func (d *Dashboard) CreateMetricWithBufSize(target string, size int) (*Metric, error) {
	return d.srv.metrics.Create(target, size)
}

// bufSizeFor takes a duration and a rate (number of data points per second)
// and returns the required ring buffer size.
// Used by CreateMetric().
func (d *Dashboard) bufSizeFor(timeRange, interval time.Duration) int {
	if interval.Nanoseconds() >= timeRange.Nanoseconds() {
		return 1
	}
	return int(timeRange.Nanoseconds() / interval.Nanoseconds())
}

// DeleteMetric deletes the metric for the given target from the server.
func (d *Dashboard) DeleteMetric(target string) error {
	return d.srv.metrics.Delete(target)
}
