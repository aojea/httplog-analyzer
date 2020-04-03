package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

// Alerter send an event based on metrics information
type Alerter interface {
	Alert(threshold int, messages chan<- string)
}

// CommonLogAlert implements the Alterter interface
type CommonLogAlert struct {
	client   client.Client
	filename string
}

// Alert depending of threshold
// TODO: Pass a context to be able to cancel the alert loop
func (c CommonLogAlert) Alert(threshold int, messages chan<- string) {
	var alert bool

	// Alert loop
	for {
		query := fmt.Sprintf("SELECT mean(value) FROM requests_total WHERE (file = '%s' AND time > now() -2m)", c.filename)
		q := client.NewQuery(query, defaultDBname, "")
		response, err := c.client.Query(q)
		if err != nil {
			log.Printf("Error querying the database: %v", err)
		}
		if err == nil && response.Error() == nil && len(response.Results[0].Series) > 0 {
			avgReq, _ := response.Results[0].Series[0].Values[0][1].(json.Number).Float64()
			// Send alert message if we cross the threshold
			if int64(avgReq) > int64(threshold) && !alert {
				messages <- fmt.Sprintf("High traffic generated an alert - hits = %v, triggered at %s", avgReq, time.Now().String())
				alert = true
			} else if int64(avgReq) < int64(threshold) && alert {
				// Restore the alert
				messages <- fmt.Sprintf("High traffic alert recovered - hits = %v, triggered at %s", avgReq, time.Now().String())
				alert = false
			}
		}
		time.Sleep(10 * time.Second)
	}
}
