package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

func TestAlert(t *testing.T) {

	config := client.HTTPConfig{}
	c, _ := client.NewHTTPClient(config)
	defer c.Close()

	var alerter Alerter
	alerter = &CommonLogAlert{client: c, filename: "test"}
	alertCh := make(chan string)
	go alerter.Alert(10, alertCh)

	cases := []struct {
		name     string
		response []byte
		message  string
	}{
		{
			name: "generate alert",
			response: []byte(`
			{"results":[
				{"statement_id":0,
				"series":
					[{
						"requests_total":["time","mean"],
						"values":[["2020-04-03T14:21:23.840000791Z",99.9]]
					}]
				}
			]}	
			`),
			message: "Alert High traffic generated an alert - hits = 99.9, triggered at 2020-04-04 01:10:43.752866 +0200 CEST m=+0.003371168",
		},
		{
			name: "recovery alert",
			response: []byte(`
			{"results":[
				{"statement_id":0,
				"series":
					[{
						"requests_total":["time","mean"],
						"values":[["2020-04-03T14:21:23.840000791Z",99.9]]
					}]
				}
			]}	
			`),
			message: "Alert High traffic generated an alert - hits = 99.9, triggered at 2020-04-04 01:10:43.752866 +0200 CEST m=+0.003371168",
		},
	}
	for _, tc := range cases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Influxdb-Version", "1.3.1")
			w.WriteHeader(http.StatusOK)
			w.Write(tc.response)
		}))
		defer ts.Close()

		query := client.Query{}
		resp, err := c.Query(query)
		if err != nil {
			t.Fatalf("unexpected error.  expected %v, actual %v", nil, err)
		}

		if got, exp := len(resp.Results), 1; got != exp {
			t.Fatalf("unexpected number of results.  expected %v, actual %v", exp, got)
		}

		select {
		case msg := <-alertCh:
			if msg != tc.message {
				t.Logf("Alert expected: %s\n Alert received: %s\n", tc.message, msg)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("Timed out waiting for msg: %s", tc.message)
		}
	}
}
