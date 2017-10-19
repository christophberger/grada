package grada

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// query is a `/query` request from Grafana.
//
// All JSON-related structs were generated from the JSON examples
// of the "SimpleJson" data source documentation
// using [JSON-to-Go](https://mholt.github.io/json-to-go/),
// with a little tweaking afterwards.
type query struct {
	PanelID int `json:"panelId"`
	Range   struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
		Raw  struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"raw"`
	} `json:"range"`
	RangeRaw struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"rangeRaw"`
	Interval   string `json:"interval"`
	IntervalMs int    `json:"intervalMs"`
	Targets    []struct {
		Target string `json:"target"`
		RefID  string `json:"refId"`
		Type   string `json:"type"`
	} `json:"targets"`
	Format        string `json:"format"`
	MaxDataPoints int    `json:"maxDataPoints"`
}

// row is used in timeseriesResponse and tableResponse.
// Grafana's JSON contains weird arrays with mixed types!
type row []interface{}

// column is used in tableResponse.
type column struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// timeseriesResponse is the response to a `/query` request
// if "Type" is set to "timeserie".
// It sends time series data back to Grafana.
type timeseriesResponse struct {
	Target     string `json:"target"`
	Datapoints []row  `json:"datapoints"`
}

// tableResponse is the response to send when "Type" is "table".
type tableResponse struct {
	Columns []column `json:"columns"`
	Rows    []row    `json:"rows"`
	Type    string   `json:"type"`
}

// ## The data aggregator

// Count is a single time series data tuple, consisting of
// a float64 value N and a timestamp T.
type Count struct {
	N float64
	T time.Time
}

// Metric is a ring buffer of Counts.
type Metric struct {
	m    sync.Mutex
	list []Count
	head int
}

// Add a single value to the ring buffer. When the ring buffer
// is full, every new value overwrites the oldest one.
func (g *Metric) Add(n float64) {
	g.m.Lock()
	g.list[g.head] = Count{n, time.Now()}
	g.head = (g.head + 1) % len(g.list)
	g.m.Unlock()
}

// Add list adds a complete Count list to the ring buffer.
func (g *Metric) AddList(c []Count) {
	g.m.Lock()
	for _, el := range c {
		g.list[g.head] = el
		g.head = (g.head + 1) % len(g.list)
	}
	g.m.Unlock()
}

// AddWithTime adds a single (value, timestamp) tuple to the ring buffer.
func (g *Metric) AppendWithTime(n float64, t time.Time) {
	g.m.Lock()
	g.list[g.head] = Count{n, t}
	g.head = (g.head + 1) % len(g.list)
	g.m.Unlock()
}

func (g *Metric) fetchMetric() *[]row {

	g.m.Lock()
	length := len(g.list)
	gcnt := make([]Count, length, length)
	head := g.head
	copy(gcnt, g.list)
	g.m.Unlock()

	rows := []row{}
	for i := 0; i < length; i++ {
		count := gcnt[(i+head)%length] // wrap around
		rows = append(rows, row{count.N, count.T.UnixNano() / 1000000})
	}
	return &rows
}

// Metrics is a map of all metric buffers, with the key being the target name.
type Metrics map[string]*Metric

// CreateMetric creates a new Metric with the given target name and buffer size
// and adds it to the Metrics map.
func (m *Metrics) CreateMetric(name string, size int) {
	(*m)[name] = &Metric{
		list: make([]Count, size, size),
	}
}

// ## The data generator

func newFakeDataFunc(max int, volatility float64) func() int {
	value := rand.Intn(max)
	return func() int {
		rnd := rand.Float64() - 0.5
		changePercent := 2 * volatility * rnd
		value += int(float64(value) * changePercent)
		return value
	}
}

// ## The server

type Server struct {
	Metrics *Metrics
}

func writeError(w http.ResponseWriter, e error, m string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("{\"error\": \"" + m + ": " + e.Error() + "\"}"))

}

func (app *Server) queryHandler(w http.ResponseWriter, r *http.Request) {
	var q bytes.Buffer

	_, err := q.ReadFrom(r.Body)
	if err != nil {
		writeError(w, err, "Cannot read request body")
		return
	}

	query := &query{}
	err = json.Unmarshal(q.Bytes(), query)
	if err != nil {
		writeError(w, err, "cannot unmarshal request body")
		return
	}

	// Our example should contain exactly one target.
	target := query.Targets[0].Target

	log.Println("Sending response for target " + target)

	// Depending on the type, we need to send either a timeseries response
	// or a table response.
	switch query.Targets[0].Type {
	case "timeserie":
		app.sendTimeseries(w, query)
	case "table":
		app.sendTable(w, query)
	}
}

func (app *Server) sendTimeseries(w http.ResponseWriter, q *query) {

	log.Println("Sending time series data")

	target := q.Targets[0].Target
	response := []timeseriesResponse{
		{
			Target:     target,
			Datapoints: (*(*app.Metrics)[target].fetchMetric()),
		},
	}

	jsonResp, err := json.Marshal(response)
	if err != nil {
		writeError(w, err, "cannot marshal timeseries response")
	}

	w.Write(jsonResp)

}

func (app *Server) sendTable(w http.ResponseWriter, q *query) {

	log.Println("Sending table data")

	response := []tableResponse{
		{
			Columns: []column{
				{Text: "Name", Type: "string"},
				{Text: "Value", Type: "number"},
				{Text: "Time", Type: "time"},
			},
			Rows: []row{
				{"Alpha", rand.Intn(100), float64(int64(time.Now().UnixNano() / 1000000))},
				{"Bravo", rand.Intn(100), float64(int64(time.Now().UnixNano() / 1000000))},
				{"Charlie", rand.Intn(100), float64(int64(time.Now().UnixNano() / 1000000))},
				{"Delta", rand.Intn(100), float64(int64(time.Now().UnixNano() / 1000000))},
			},
			Type: "table",
		},
	}

	jsonResp, err := json.Marshal(response)
	if err != nil {
		writeError(w, err, "cannot marshal table response")
	}

	w.Write(jsonResp)

}

// A search request from Grafana expects a list of target names as a response.
// These names are shown in the metrics dropdown when selecting a metric in
// the Metrics tab of a panel.
func (a *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	var targets []string
	for t, _ := range *(a.Metrics) {
		targets = append(targets, t)
	}
	resp, err := json.Marshal(targets)
	if err != nil {
		writeError(w, err, "cannot marshal targets response")
	}
	w.Write(resp)
}

func StartServer() {

	app := &Server{Metrics: &Metrics{}}

	// Grafana expects a "200 OK" status for "/" when testing the connection.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/query", app.queryHandler)

	// Start the server.
	log.Println("start grafanago")
	defer log.Println("stop grafanago")
	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatalln(err)
	}
}
