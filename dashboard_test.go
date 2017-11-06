package grada

import (
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDashboard_CreateMetricWithBufSize(t *testing.T) {
	type args struct {
		target string
		size   int
	}

	mt := &metrics{sync.Mutex{}, map[string]*Metric{}}

	tests := []struct {
		name    string
		metrics *metrics
		args    args
		wantErr bool
	}{
		{
			"create1",
			mt,
			args{"target1", 10},
			false,
		},
		{
			"create1again",
			mt,
			args{"target1", 10},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dashboard{
				srv: &server{
					metrics: tt.metrics,
				},
			}
			got, err := d.CreateMetricWithBufSize(tt.args.target, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.CreateMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := d.srv.metrics.metric[tt.args.target]
			if !cmp.Equal(got, want, cmp.AllowUnexported((*got), (*got).m)) {
				t.Errorf("Server.CreateMetric():\ngot  %v\nwant %v\ndiff:\n%s", got, want, cmp.Diff(got, want, cmp.AllowUnexported(*got, (*got).m)))
			}
		})
	}
}

func TestGetDashboard(t *testing.T) {
	tests := []struct {
		name string
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDashboard()
			if got.srv == nil {
				t.Errorf("GetDashboard().srv == nil")
			}
			if got.srv.metrics == nil {
				t.Errorf("GetDashboard().srv.metrics == nil")
			}
		})
	}
}

func TestDashboard_bufSizeFor(t *testing.T) {
	tests := []struct {
		name                string
		timeRange, interval time.Duration
		want                int
	}{
		{"1min, 1s", time.Minute, time.Second, 60},
		{"1h, 10s", time.Hour, 10 * time.Second, 360},
		{"12s, 11s", 12 * time.Second, 11 * time.Second, 1},
		{"1min, 2min", time.Minute, 2 * time.Minute, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dashboard{}
			if got := d.bufSizeFor(tt.timeRange, tt.interval); got != tt.want {
				t.Errorf("Dashboard.For() = %v, want %v", got, tt.want)
			}
		})
	}
}
