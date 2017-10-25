package grada

type Dashboard struct {
	Server
}

// CreateMetric creates a new metric for the given target and with the
// given buffer size, and stores this metric in the server.
func (d *Dashboard) CreateMetric(target string, size int) (*Metric, error) {
	return d.Metrics.Create(target, size)
}

// DeleteMetric deletes the metric for the given target from the server.
func (d *Dashboard) DeleteMetric(target string) error {
	return d.Metrics.Delete(target)
}
