package grada

type Dashboard struct {
	srv *server
}

var d *Dashboard

// GetDashboard initializes and/or returns the only existing dashboard
func GetDashboard() *Dashboard {
	if d == nil {
		d = &Dashboard{
			srv: &server{
				metrics: &metrics{},
			},
		}
		d.srv = startServer()
	}
	return d
}

// CreateMetric creates a new metric for the given target and with the
// given buffer size, and stores this metric in the server.
// Creating a metric for an existing target is an error. To replace a metric
// (which is rarely needed), call DeleteMetric first.
func (d *Dashboard) CreateMetric(target string, size int) (*Metric, error) {
	return d.srv.metrics.Create(target, size)
}

// DeleteMetric deletes the metric for the given target from the server.
func (d *Dashboard) DeleteMetric(target string) error {
	return d.srv.metrics.Delete(target)
}
