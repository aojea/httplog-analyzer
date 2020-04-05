package main

import (
	"encoding/json"
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
	client   client.Client
	filename string
}

// NewCommonLogDisplay returns a new CommonLogDisplay
func NewCommonLogDisplay(c client.Client, f string) *CommonLogDisplay {
	return &CommonLogDisplay{
		client:   c,
		filename: f,
	}
}

// Display metrics from graphite
func (c CommonLogDisplay) Display(eventCh <-chan string) error {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	uiEvents := ui.PollEvents()

	// Paragraph
	p := widgets.NewParagraph()
	p.Title = "HTTP Log Analyzer"
	p.Text = "\n\n     PRESS q TO QUIT"
	p.SetRect(0, 0, 80, 5)

	// Create list of alerts
	alerts := widgets.NewList()
	alerts.Title = "Alert History"

	alerts.TextStyle = ui.NewStyle(ui.ColorYellow)
	alerts.WrapText = false
	alerts.SetRect(0, 5, 80, 15)

	// Create list of top sections
	l := widgets.NewList()
	l.Title = "Top Visited sections"

	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.SetRect(0, 15, 30, 25)

	// Create Plot with the bytes received
	p1 := widgets.NewPlot()
	p1.Title = "KiloBytes per second"
	p1.Marker = widgets.MarkerDot
	p1.Data = make([][]float64, 2)
	p1.SetRect(30, 15, 55, 25)
	p1.AxesColor = ui.ColorWhite
	p1.LineColors[0] = ui.ColorCyan
	p1.PlotType = widgets.ScatterPlot

	// Create Plot with the requests received
	p2 := widgets.NewPlot()
	p2.Title = "Requests per second"
	p2.Marker = widgets.MarkerDot
	p2.Data = make([][]float64, 2)
	p2.SetRect(55, 15, 80, 25)
	p2.AxesColor = ui.ColorWhite
	p2.LineColors[0] = ui.ColorCyan
	p2.PlotType = widgets.ScatterPlot

	// Draw
	draw := func() {
		// update dashboard every 10 second
		l.Rows = c.getTopSection()
		p1.Data[0] = c.getBytesSecond()
		p2.Data[0] = c.getRequestsSecond()
		ui.Render(p, l, alerts, p1, p2)
	}
	draw()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return nil
			}
		case msg := <-eventCh:
			alerts.Rows = append(alerts.Rows, msg)
			// Show only last 10 alert events
			if len(alerts.Rows) > 10 {
				alerts.Rows = alerts.Rows[1:]
			}
			draw()

		case <-time.After(10 * time.Second):
			draw()
		}
	}

	return nil
}

// TODO: getTopSection, getBytesSecond and getBytesSecond can reduce a lot of duplicate code
func (c CommonLogDisplay) getTopSection() []string {
	var result []string
	// Statsd aggregates by 10s, otherwise we should do a subquery to aggregate first all occurrences during the period
	query := fmt.Sprintf("SELECT top(value,5),section FROM requests_section_count WHERE (file = '%s' AND time > now() - 15s)", c.filename)
	q := client.NewQuery(query, "statsd", "")
	response, err := c.client.Query(q)
	if err == nil && response.Error() == nil && len(response.Results[0].Series) > 0 {
		for i := range response.Results[0].Series[0].Values {
			// t := response.Results[0].Series[0].Values[i][0]
			top := response.Results[0].Series[0].Values[i][1]
			section := response.Results[0].Series[0].Values[i][2]
			// termui doesn't support \t tabs :/
			result = append(result, fmt.Sprintf("[%v] %v  Hits: %v", i+1, section, top))
		}
	}
	return result
}

func (c CommonLogDisplay) getBytesSecond() []float64 {
	var result []float64
	// Get last 2 minutes bytes time series
	query := fmt.Sprintf("SELECT last(value) FROM requests_bytes_count WHERE (file = '%s' AND time > now() - 2m) GROUP BY time(10s) FILL(0)", c.filename)
	q := client.NewQuery(query, "statsd", "")
	response, err := c.client.Query(q)
	if err == nil && response.Error() == nil && len(response.Results[0].Series) > 0 {
		for i := range response.Results[0].Series[0].Values {
			// t := response.Results[0].Series[0].Values[i][0]
			bytes, _ := response.Results[0].Series[0].Values[i][1].(json.Number).Float64()
			// TODO: Aggregated period is a constant (10s), make it a variable
			kbps := bytes / (1024 * 10)
			result = append(result, kbps)
		}
		// TODO: review influxdb query to not return always last result as 0
		result = result[:len(result)-2]
	}
	return result
}

func (c CommonLogDisplay) getRequestsSecond() []float64 {
	var result []float64
	// Get last 2 mins requests time series
	query := fmt.Sprintf("SELECT last(value) FROM requests_total WHERE (file = '%s' AND time > now() - 2m) GROUP BY time(10s) FILL(0)", c.filename)
	q := client.NewQuery(query, "statsd", "")
	response, err := c.client.Query(q)
	if err == nil && response.Error() == nil && len(response.Results[0].Series) > 0 {
		for i := range response.Results[0].Series[0].Values {
			// t := response.Results[0].Series[0].Values[i][0]
			requests, _ := response.Results[0].Series[0].Values[i][1].(json.Number).Float64()
			// TODO: Aggregated period is a constant (10s), make it a variable
			rps := requests / 10
			result = append(result, rps)
		}
	}
	return result
}
