package main

import (
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

// NewCommonLogDisplay returns a new CommonLogDisplay
func NewCommonLogDisplay(c client.Client) *CommonLogDisplay {
	return &CommonLogDisplay{
		client: c,
	}
}

// Display metrics from graphite
func (c CommonLogDisplay) Display(eventCh <-chan string) error {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	ev := ui.PollEvents()

	// Paragraph
	p := widgets.NewParagraph()
	p.Text = "HTTP Log Analyzer"
	p.SetRect(0, 0, 25, 5)

	// Create list of top sections
	l := widgets.NewList()
	l.Title = "Top Visited sections"

	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.SetRect(0, 0, 25, 8)

	// Create list of alerts
	alerts := widgets.NewList()
	alerts.Title = "Alert History"

	alerts.TextStyle = ui.NewStyle(ui.ColorYellow)
	alerts.WrapText = false
	alerts.SetRect(0, 8, 25, 8)

	// Create SparkLine with the bytes received
	data := []float64{4, 2, 1, 6, 3, 9, 1, 4, 2, 15, 14, 9, 8, 6, 10, 13, 15, 12, 10, 5, 3, 6, 1, 7, 10, 10, 14, 13, 6}

	sl0 := widgets.NewSparkline()
	sl0.Data = data[3:]
	sl0.LineColor = ui.ColorGreen

	// single
	slg0 := widgets.NewSparklineGroup(sl0)
	slg0.Title = "Sparkline 0"
	slg0.SetRect(0, 0, 20, 10)

	// Create SparkLine with the number of requests

	sl1 := widgets.NewSparkline()
	sl1.Data = data[3:]
	sl1.LineColor = ui.ColorGreen

	// single
	slg1 := widgets.NewSparklineGroup(sl1)
	slg1.Title = "Sparkline 1"
	slg1.SetRect(0, 0, 20, 10)

	for {
		select {
		case e := <-ev:
			switch e.Type {
			case ui.KeyboardEvent:
				// quit on any keyboard event
				return nil
			case ui.ResizeEvent:
			}
		case msg := <-eventCh:
			alerts.Rows = append(alerts.Rows, msg)
			// Show only last 10 alert events
			if len(alerts.Rows) > 10 {
				alerts.Rows = alerts.Rows[1:]
			}
			ui.Render(p, l, alerts, slg0, slg1)

		case <-time.After(10 * time.Second):
			// update dashboard every 10 second
			l.Rows = c.getTopSection()
			sl0.Data = c.getBytesSecond()
			sl1.Data = c.getRequestsSecond()
			ui.Render(p, l, alerts, slg0, slg1)
		}
	}

	return nil
}

func (c CommonLogDisplay) getTopSection() []string {
	q := client.NewQuery("SELECT top(f3), f3 FROM m ", "statsd", "")
	response, err := c.client.Query(q)
	if err != nil {
		return nil
	}
	if err == nil && response.Error() == nil && len(response.Results[0].Series) > 0 {
		// _ = response.Results[0].Series[0]
		return nil
	}
	return nil
}

func (c CommonLogDisplay) getBytesSecond() []float64 {
	return nil
}

func (c CommonLogDisplay) getRequestsSecond() []float64 {
	return nil
}
