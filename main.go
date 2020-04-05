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
const defaultDBname = "statsd"

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
	// so we can use different log formats
	logParser := NewCommonLog(c, filepath.Base(*file))

	// Open file
	t, err := tail.TailFile(*file, tail.Config{
		// start at the end of the file
		Location:  &tail.SeekInfo{1, os.SEEK_END},
		Poll:      true,
		Follow:    true,
		ReOpen:    false,
		MustExist: true})
	if err != nil {
		log.Fatal(err)
	}

	defer t.Stop()
	defer t.Cleanup()

	// Process alerts
	alerter := NewAlert(i, filepath.Base(*file))
	alertCh := make(chan string)
	go alerter.Alert(*threshold, alertCh)

	// Process file
	go func() {
		for line := range t.Lines {
			err := logParser.LogParse(line.Text)
			if err != nil {
				log.Println(err)
			}

		}

	}()

	// Display metrics
	displayer := NewCommonLogDisplay(i, filepath.Base(*file))
	// Displayer blocks until "q" is pressed
	displayer.Display(alertCh)
}
