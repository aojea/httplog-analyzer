package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/hpcloud/tail"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

const defaultFile = "/tmp/access.log"

func main() {
	// Configuration
	file := flag.String("f", defaultFile, "log file")
	statsdAddress := flag.String("s", "127.0.0.1:8125", "Statsd server address")
	influxAddress := flag.String("i", "http://127.0.0.1:8086", "InfluxDB server address")
	threshold := flag.Int("t", 10, "Threshold requests per second averaged over a 2 minutes slot")
	help := flag.String("h", "", "help")
	flag.Parse()
	if len(os.Args) > 8 || len(*help) > 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create a statsd client
	c, err := statsd.New(*statsdAddress)
	if err != nil {
		log.Fatalf("Error creating StatsD client: %v", err)
	}
	c.Namespace = filepath.Base(*file)
	defer c.Close()

	// Create a influxdb client
	i, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: *influxAddress,
	})
	if err != nil {
		log.Fatalf("Error creating InfluxDB Client: %v", err)
	}
	defer i.Close()
	// Create a LogProcessor
	// Using an interface allows to replace the log processor
	// for different log formats
	var logParser LogParser
	logParser = CommonLog{}

	// Open file
	t, err := tail.TailFile(*file, tail.Config{
		Poll:      true,
		Follow:    true,
		ReOpen:    false,
		MustExist: true})
	if err != nil {
		log.Fatal(err)
	}

	defer t.Stop()
	defer t.Cleanup()

	// Display metrics
	var displayer Displayer
	displayer = CommonLogDisplay{}
	go displayer.Display(i)

	// Alert
	var alerter Alerter
	alerter = CommonLogAlert{}
	go func() {
		status := false
		for {
			alert, err := alerter.Alert(i, *threshold)
			if err != nil {
				log.Printf("Error with the alerting subsystem: %v", err)
				return
			}
			if alert != status {
				if alert {
					c.SimpleEvent("High traffic generated an alert", "hits = {value}, triggered at {time}”")
				} else {
					c.SimpleEvent("High traffic alert recovered", "hits = {value}, triggered at {time}”")
				}

			}
		}
	}()

	// Process file
	for line := range t.Lines {
		err := logParser.LogParse(c, line.Text)
		if err != nil {
			log.Println(err)
		}
	}
}
