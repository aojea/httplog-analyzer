package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

// &{[{0 [{requests_total map[] [time mean] [[2020-04-03T14:21:23.840000791Z 99.9]] false}] [] }] }

func TestAlert_Alert(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := client.Response{
			Results: []client.Result{{
				StatementId: 0,
				Series:      nil,
				Messages:    nil,
				Err:         "",
			}},
			Err: "",
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Influxdb-Version", "1.3.1")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		_ = enc.Encode(data)
	}))
	defer ts.Close()

	config := client.HTTPConfig{Addr: ts.URL}
	c, _ := client.NewHTTPClient(config)
	defer c.Close()

	query := client.Query{}
	resp, err := c.Query(query)
	if err != nil {
		t.Fatalf("unexpected error.  expected %v, actual %v", nil, err)
	}

	if got, exp := len(resp.Results), 1; got != exp {
		t.Fatalf("unexpected number of results.  expected %v, actual %v", exp, got)
	}

	var alerter Alerter
	alerter = &CommonLogAlert{client: c, filename: "test"}
	alertCh := make(chan string)
	go alerter.Alert(10, alertCh)

	select {
	case msg := <-alertCh:
		fmt.Println("Alert", msg)
	case <-time.After(3 * time.Second):
		fmt.Println("timeout 2")
	}
}
