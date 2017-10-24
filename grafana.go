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
	"errors"
	"log"
	"math/rand"
	"net/http"
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

// ## The server

type Server struct {
	Metrics *Metrics
}

func writeError(w http.ResponseWriter, e error, m string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("{\"error\": \"" + m + ": " + e.Error() + "\"}"))

}

func (srv *Server) queryHandler(w http.ResponseWriter, r *http.Request) {
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
		srv.sendTimeseries(w, query)
	case "table":
		srv.sendTable(w, query)
	}
}

func (srv *Server) sendTimeseries(w http.ResponseWriter, q *query) {

	log.Println("Sending time series data")

	target := q.Targets[0].Target
	metric, ok := srv.Metrics.metric[target]
	if !ok {
		writeError(w, errors.New("No metric for target "+target), "")
	}
	response := []timeseriesResponse{
		{
			Target:     target,
			Datapoints: *(metric.fetchDatapoints()),
		},
	}

	jsonResp, err := json.Marshal(response)
	if err != nil {
		writeError(w, err, "cannot marshal timeseries response")
	}

	w.Write(jsonResp)

}

func (srv *Server) sendTable(w http.ResponseWriter, q *query) {

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
func (srv *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	var targets []string
	for t, _ := range srv.Metrics.metric {
		targets = append(targets, t)
	}
	resp, err := json.Marshal(targets)
	if err != nil {
		writeError(w, err, "cannot marshal targets response")
	}
	w.Write(resp)
}

func StartServer() *Server {

	server := &Server{Metrics: &Metrics{}}

	// Grafana expects a "200 OK" status for "/" when testing the connection.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/query", server.queryHandler)
	http.HandleFunc("/search", server.searchHandler)

	// Start the server.
	log.Println("start grafanago")
	defer log.Println("stop grafanago")
	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatalln(err)
	}
	return server
}
