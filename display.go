package main

import (
	"fmt"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

// Displayer displays the metrics information
type Displayer interface {
	Display(i client.Client) error
}

// CommonLogDisplay implements the Displayer interface
type CommonLogDisplay struct{}

// Display metrics from graphite
func (c CommonLogDisplay) Display(i client.Client) error {
	q := client.NewQuery("SELECT count(value) FROM cpu_load", "mydb", "")
	if response, err := i.Query(q); err == nil && response.Error() == nil {
		fmt.Println(response.Results)
	}
	return nil
}
