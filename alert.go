package main

import (
	"fmt"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

// Alerter send an event based on metrics information
type Alerter interface {
	Alert(i client.Client, threshold int) (bool, error)
}

// CommonLogAlert implements the Alterter interface
type CommonLogAlert struct{}

// Alert depending of threshold
func (c CommonLogAlert) Alert(i client.Client, threshold int) (bool, error) {
	q := client.NewQuery("SELECT count(value) FROM cpu_load", "mydb", "")
	response, err := i.Query(q)
	if err != nil {
		return false, err
	}
	if err == nil && response.Error() == nil {
		fmt.Println(response.Results)
		return true, nil
	}
	return false, nil
}
