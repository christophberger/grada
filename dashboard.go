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

// CreateMetric creates a new metric for the given target and with the
// given buffer size, and stores this metric in the server.
// Creating a metric for an existing target is an error. To replace a metric
// (which is rarely needed), call DeleteMetric first.
func (d *Dashboard) CreateMetric(target string, size int) (*Metric, error) {
	return d.srv.metrics.Create(target, size)
}

// BufSizeFor takes a duration and a rate (number of data points per second)
// and returns the required ring buffer size.
// BufSizeFor use with Metrics.Create() as follows:
//         d.CreateMetric("mytarget", d.BufSizeFor(5* time.Minute, 10 * time.Second))
func (d *Dashboard) BufSizeFor(timeRange, interval time.Duration) int {
	if interval.Nanoseconds() >= timeRange.Nanoseconds() {
		return 1
	}
	return int(timeRange.Nanoseconds() / interval.Nanoseconds())
}

// DeleteMetric deletes the metric for the given target from the server.
func (d *Dashboard) DeleteMetric(target string) error {
	return d.srv.metrics.Delete(target)
}
