package grada

import (
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestMetric_Add(t *testing.T) {
	type fields struct {
		list []Count
		head int
	}
	type args struct {
		n float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		newHead int
	}{
		{
			name: "target1",
			fields: fields{
				list: []Count{{1, time.Now()}, {2, time.Now()}, {3, time.Now()}},
				head: 1},
			args:    args{n: 4},
			newHead: 2,
		},

		{
			name: "target2",
			fields: fields{
				list: []Count{{4, time.Now()}, {5, time.Now()}, {6, time.Now()}},
				head: 2},
			args:    args{n: 7},
			newHead: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Metric{
				m:    sync.Mutex{},
				list: tt.fields.list,
				head: tt.fields.head,
			}
			g.Add(tt.args.n)
			if tt.fields.list[tt.fields.head].N != tt.args.n {
				t.Errorf("failed adding %f to metric for target %s", tt.args.n, tt.name)
			}
		})
	}
}

func TestMetric_AddWithTime(t *testing.T) {
	type fields struct {
		list []Count
		head int
	}
	type args struct {
		n float64
		t time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		newHead int
	}{
		{
			name: "target1",
			fields: fields{
				list: []Count{{1, time.Now()}, {2, time.Now()}, {3, time.Now()}},
				head: 1},
			args:    args{n: 4, t: time.Date(2017, time.October, 25, 11, 16, 54, 0, time.UTC)},
			newHead: 2,
		},

		{
			name: "target2",
			fields: fields{
				list: []Count{{4, time.Now()}, {5, time.Now()}, {6, time.Now()}},
				head: 2},
			args:    args{n: 7, t: time.Date(2017, time.October, 25, 11, 16, 54, 0, time.UTC)},
			newHead: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Metric{
				m:    sync.Mutex{},
				list: tt.fields.list,
				head: tt.fields.head,
			}
			g.AddWithTime(tt.args.n, tt.args.t)
			if tt.fields.list[tt.fields.head].N != tt.args.n {
				t.Errorf("failed adding %f to metric for target %s", tt.args.n, tt.name)
			}
			if tt.fields.list[tt.fields.head].T != tt.args.t {
				t.Errorf("failed adding time %s to metric for target %s", tt.args.t.String(), tt.name)
			}
		})
	}
}

func TestMetric_AddCount(t *testing.T) {
	type fields struct {
		list []Count
		head int
	}
	type args struct {
		c Count
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		newHead int
	}{
		{
			name: "target1",
			fields: fields{
				list: []Count{{1, time.Now()}, {2, time.Now()}, {3, time.Now()}},
				head: 1},
			args:    args{c: Count{N: 4, T: time.Date(2017, time.October, 25, 11, 16, 54, 0, time.UTC)}},
			newHead: 2,
		},

		{
			name: "target2",
			fields: fields{
				list: []Count{{4, time.Now()}, {5, time.Now()}, {6, time.Now()}},
				head: 2},
			args:    args{c: Count{N: 7, T: time.Date(2017, time.October, 25, 11, 16, 54, 0, time.UTC)}},
			newHead: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Metric{
				m:    sync.Mutex{},
				list: tt.fields.list,
				head: tt.fields.head,
			}
			g.AddCount(tt.args.c)
			if got := tt.fields.list[tt.fields.head].N; got != tt.args.c.N {
				t.Errorf("AddCount(%f, %s) failed for %s", tt.args.c.N, tt.args.c.T.String(), tt.name, got)
			}
			if got := tt.fields.list[tt.fields.head].T; got != tt.args.c.T {
				t.Errorf("AddCount(%f, %s) failed for %s - got ", tt.args.c.N, tt.args.c.T.String(), tt.name, got)
			}
		})
	}
}

func TestMetric_fetchDatapoints(t *testing.T) {
	type fields struct {
		list []Count
		head int
	}

	t1 := time.Date(2017, time.October, 25, 11, 16, 54, 0, time.UTC)
	t2 := time.Date(2017, time.October, 25, 11, 17, 54, 0, time.UTC)
	t3 := time.Date(2017, time.October, 25, 11, 18, 54, 0, time.UTC)
	t1ms := t1.UnixNano() / 1000000
	t2ms := t2.UnixNano() / 1000000
	t3ms := t3.UnixNano() / 1000000

	tests := []struct {
		name   string
		fields fields
		want   *[]row
	}{
		{
			"fetch1",
			fields{[]Count{{3, t3}, {1, t1}, {2, t2}}, 1},
			&[]row{{1.0, t1ms}, {2.0, t2ms}, {3.0, t3ms}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Metric{
				m:    sync.Mutex{},
				list: tt.fields.list,
				head: tt.fields.head,
			}
			if got := g.fetchDatapoints(); !cmp.Equal(got, tt.want) {
				t.Errorf("Metric.fetchDatapoints():\ngot  %#v,\nwant %#v\nDiff: %s", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestMetrics_Get(t *testing.T) {
	type fields struct {
		metric map[string]*Metric
	}
	type args struct {
		target string
	}

	t1 := time.Date(2017, time.October, 25, 11, 16, 54, 0, time.UTC)
	t2 := time.Date(2017, time.October, 25, 11, 17, 54, 0, time.UTC)
	t3 := time.Date(2017, time.October, 25, 11, 18, 54, 0, time.UTC)

	metric := &Metric{sync.Mutex{}, []Count{{3, t3}, {1, t1}, {2, t2}}, 1}

	tests := []struct {
		name    string
		fields  *fields
		args    args
		want    *Metric
		wantErr bool
	}{
		{
			"get1",
			&fields{map[string]*Metric{"target1": metric}},
			args{"target1"},
			metric,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &Metrics{
				m:      sync.Mutex{},
				metric: tt.fields.metric,
			}
			got, err := metrics.Get(tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Metrics.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want { // strict identity required
				t.Errorf("Metrics.Get():\ngot  %v\n want %v", &got, &tt.want)
			}
		})
	}
}

func TestMetrics_Put(t *testing.T) {
	type fields struct {
		metric map[string]*Metric
	}
	type args struct {
		target string
		metric *Metric
	}
	t1 := time.Date(2017, time.October, 25, 11, 16, 54, 0, time.UTC)
	t2 := time.Date(2017, time.October, 25, 11, 17, 54, 0, time.UTC)
	t3 := time.Date(2017, time.October, 25, 11, 18, 54, 0, time.UTC)
	metric := &Metric{sync.Mutex{}, []Count{{3, t3}, {1, t1}, {2, t2}}, 1}

	tests := []struct {
		name    string
		fields  *fields
		args    args
		wantErr bool
	}{
		{
			"put1",
			&fields{map[string]*Metric{"target1": metric}},
			args{"target1", metric},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				m:      sync.Mutex{},
				metric: map[string]*Metric{},
			}
			err := m.Put(tt.args.target, tt.args.metric)
			if (err != nil) != tt.wantErr {
				t.Errorf("Metrics.Put() error = %v, wantErr %v", err, tt.wantErr)
			}
			if mt, err := m.Get(tt.args.target); err != nil || mt != metric {
				t.Errorf("Metrics.Put():\ngot  %v\nwant %v", &mt, &metric)
			}
		})
	}
}

func TestMetrics_Delete(t *testing.T) {
	type fields struct {
		metric map[string]*Metric
	}
	type args struct {
		target string
	}
	t1 := time.Date(2017, time.October, 25, 11, 16, 54, 0, time.UTC)
	t2 := time.Date(2017, time.October, 25, 11, 17, 54, 0, time.UTC)
	t3 := time.Date(2017, time.October, 25, 11, 18, 54, 0, time.UTC)
	metric := &Metric{sync.Mutex{}, []Count{{3, t3}, {1, t1}, {2, t2}}, 1}

	tests := []struct {
		name    string
		fields  *fields
		args    args
		wantErr bool
	}{
		{
			"delete1",
			&fields{map[string]*Metric{"target1": metric}},
			args{"target1"},
			false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				m:      sync.Mutex{},
				metric: tt.fields.metric,
			}
			if err := m.Delete(tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("Metrics.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(tt.fields.metric) > 0 {
				t.Errorf("Metrics.Delete():\ngot  %v\nwant <nil>", m.metric[tt.args.target])
			}
		})
	}
}

func TestMetrics_Create(t *testing.T) {
	type args struct {
		target string
		size   int
	}

	metric := map[string]*Metric{}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"metric1",
			args{"target1", 1000},
			false,
		},
		{
			"metric1again",
			args{"target1", 1000},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &Metrics{
				m:      sync.Mutex{},
				metric: metric,
			}
			got, err := metrics.Create(tt.args.target, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("Metrics.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := metrics.metric[tt.args.target]
			if got != want {
				t.Errorf("Metrics.Create() = %v, want %v", got, want)
			}
			if cap(got.list) != tt.args.size {
				t.Errorf("Metrics.Create(): got size %d, want %d", cap(got.list), tt.args.size)
			}
		})
	}
}