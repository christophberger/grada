package grada

import (
	"sync"
	"testing"
)

func TestServer_CreateMetric(t *testing.T) {
	type args struct {
		target string
		size   int
	}

	metrics := &Metrics{sync.Mutex{}, map[string]*Metric{}}

	tests := []struct {
		name    string
		Metrics *Metrics
		args    args
		wantErr bool
	}{
		{
			"create1",
			metrics,
			args{"target1", 1000},
			false,
		},
		{
			"create1again",
			metrics,
			args{"target1", 1000},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				Metrics: tt.Metrics,
			}
			got, err := srv.CreateMetric(tt.args.target, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.CreateMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := srv.Metrics.metric[tt.args.target]
			if got != want { // strict identity
				t.Errorf("Server.CreateMetric() = %v, want %v", got, want)
			}
		})
	}
}
