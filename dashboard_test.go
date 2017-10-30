package grada

import (
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDashboard_CreateMetric(t *testing.T) {
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
			got, err := d.CreateMetric(tt.args.target, tt.args.size)
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
