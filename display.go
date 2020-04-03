package main

import (
	"fmt"
	"log"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

// Displayer displays the metrics information
type Displayer interface {
	Display() error
}

// CommonLogDisplay implements the Displayer interface
type CommonLogDisplay struct {
	client client.Client
}

// Display metrics from graphite
func (c CommonLogDisplay) Display() error {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	ev := ui.PollEvents()
	tick := time.Tick(time.Second)

	for {
		select {
		case e := <-ev:
			switch e.Type {
			case ui.KeyboardEvent:
				// quit on any keyboard event
				return nil
			case ui.ResizeEvent:
			}
		case <-tick:
			// update dashboard every second
			p := widgets.NewParagraph()
			p.Text = "Hello World!"
			p.SetRect(0, 0, 25, 5)

			ui.Render(p)
		}
	}

	q := client.NewQuery("SELECT count(value) FROM cpu_load", "mydb", "")
	response, err := c.client.Query(q)
	if err != nil {
		return err
	}
	if err == nil && response.Error() == nil {
		fmt.Println(response.Results)

	}
	return nil
}

bps := fmt.Sprintf("SELECT mean(value)  / 10 FROM request_bytes_count WHERE (file = '%s' AND time > now() -2m)", c.filename)
topSection := bps := fmt.Sprintf("SELECT mean(value)  / 10 FROM request_bytes_count WHERE (file = '%s' AND time > now() -2m)", c.filename)