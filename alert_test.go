package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

// TestAlert tests the alert logic:
// 1. Create an alert
// 2. Check that it doesn't retrigger the alert
// 3. Recover the alert
// 4. Check that it doesn't retrigger the recovery
// The default alert poll period is 10 sec so the test has to wait it times out
// and takes more than 35 seconds to be executed
func TestAlert(t *testing.T) {
	var response []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Influxdb-Version", "1.3.1")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}))
	defer ts.Close()

	config := client.HTTPConfig{Addr: ts.URL}
	c, _ := client.NewHTTPClient(config)
	defer c.Close()

	alerter := NewAlert(c, "test")
	alertCh := make(chan string)
	go alerter.Alert(10, alertCh)
	defer close(alertCh)

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
			message: "Alert High traffic generated an alert - hits = 99.9, triggered at",
		},
		{
			name: "alert high, do nothing",
			response: []byte(`
			{"results":[
				{"statement_id":0,
				"series":
					[{
						"requests_total":["time","mean"],
						"values":[["2020-04-03T14:21:23.840000791Z",19.9]]
					}]
				}
			]}	
			`),
			message: "",
		},
		{
			name: "recovery alert",
			response: []byte(`
			{"results":[
				{"statement_id":0,
				"series":
					[{
						"requests_total":["time","mean"],
						"values":[["2020-04-03T14:21:23.840000791Z",5.1]]
					}]
				}
			]}	
			`),
			message: "High traffic alert recovered - hits = 5.1, triggered at",
		},
		{
			name: "no alert, do nothing",
			response: []byte(`
			{"results":[
				{"statement_id":0,
				"series":
					[{
						"requests_total":["time","mean"],
						"values":[["2020-04-03T14:21:23.840000791Z",3.1]]
					}]
				}
			]}	
			`),
			message: "",
		},
	}
	// Has to be serial, order of execution matters
	for _, tc := range cases {
		response = tc.response
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
			if strings.Contains(msg, tc.message) {
				t.Logf("Alert expected: %s\n Alert received: %s\n", tc.message, msg)
			}
		case <-time.After(15 * time.Second):
			if len(tc.message) > 0 {
				t.Fatalf("Timed out waiting for msg: %s", tc.message)
			}
		}
	}
}
