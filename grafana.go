package grada

// Code required for communicating with Grafana:
// * Server
// * Handlers
// * Structs
//
// Grafana sends three queries:
// * /search for retrieving the available targets
// * /query for requesting new sets of data
// * /annotation for requesting chart annotations

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
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

var debug bool

// ## The server

// server is a Web API server for Grafana. It manages a list of metrics
// by target name. When Grafana requests new data for a target,
// the server returns the current list of metrics for that target.
type server struct {
	metrics *metrics
}

func writeError(w http.ResponseWriter, e error, m string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("{\"error\": \"" + m + ": " + e.Error() + "\"}"))

}

func (srv *server) queryHandler(w http.ResponseWriter, r *http.Request) {
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

	// Depending on the type, we need to send either a timeseries response
	// or a table response.
	switch query.Targets[0].Type {
	case "timeserie":
		srv.sendTimeseries(w, query)
	case "table":
		srv.sendTable(w, query)
	}
}

// sendTimeseries creates and writes a JSON response to a request for time series data.
func (srv *server) sendTimeseries(w http.ResponseWriter, q *query) {

	response := []timeseriesResponse{}

	for _, t := range q.Targets {
		target := t.Target
		metric, err := srv.metrics.Get(target)
		if err != nil {
			writeError(w, err, "Cannot get metric for target "+target)
			return
		}
		response = append(response, timeseriesResponse{
			Target:     target,
			Datapoints: *(metric.fetchDatapoints(q.Range.From, q.Range.To, q.MaxDataPoints)),
		})
	}

	jsonResp, err := json.Marshal(response)
	if err != nil {
		writeError(w, err, "cannot marshal timeseries response")
	}

	w.Write(jsonResp)

}

// TODO: Just a dummy for now
// sendTable creates and writes a JSON response to a request for table data
func (srv *server) sendTable(w http.ResponseWriter, q *query) {

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
func (srv *server) searchHandler(w http.ResponseWriter, r *http.Request) {
	var targets []string
	for t, _ := range srv.metrics.metric {
		targets = append(targets, t)
	}
	resp, err := json.Marshal(targets)
	if err != nil {
		writeError(w, err, "cannot marshal targets response")
	}
	w.Write(resp)
}

// startServer creates and starts the API server.
func startServer() *server {

	server := &server{
		metrics: &metrics{
			metric: map[string]*Metric{},
		},
	}

	// Grafana expects a "200 OK" status for "/" when testing the connection.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/query", server.queryHandler)
	http.HandleFunc("/search", server.searchHandler)

	// Determine the port. Default is 3001 but can be changed via
	// environment variable GRADA_PORT.
	port := "3001"
	portenv := os.Getenv("GRADA_PORT")
	if portenv != "" {
		port = portenv
	}

	// Start the server.
	go http.ListenAndServe(":"+port, nil)
	return server
}
