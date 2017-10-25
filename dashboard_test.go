package grada

import (
	"sync"
	"testing"
)

func TestDashboard_CreateMetric(t *testing.T) {
	type args struct {
		target string
		size   int
	}

	mt := &metrics{sync.Mutex{}, map[string]*Metric{}}

	tests := []struct {
		name    string
		Metrics *metrics
		args    args
		wantErr bool
	}{
		{
			"create1",
			mt,
			args{"target1", 1000},
			false,
		},
		{
			"create1again",
			mt,
			args{"target1", 1000},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Dashboard{
				s: &server{
					metrics: tt.Metrics,
				},
			}
			got, err := srv.CreateMetric(tt.args.target, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.CreateMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := d.s.metrics.metric[tt.args.target]
			if got != want { // strict identity
				t.Errorf("Server.CreateMetric() = %v, want %v", got, want)
			}
		})
	}
}
