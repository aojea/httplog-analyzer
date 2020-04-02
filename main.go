package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/DataDog/datadog-go/statsd"
	graphite "github.com/JensRantil/graphite-client"
	"github.com/hpcloud/tail"
)

const defaultFile = "/tmp/access.log"

func main() {
	// Configuration
	file := flag.String("f", defaultFile, "log file")
	statsdAddress := flag.String("s", "127.0.0.1:8125", "Statsd server address")
	graphiteAddress := flag.String("g", "http://localhost:8080/render/", "Graphite server address")
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
		log.Fatal(err)
	}
	c.Namespace = filepath.Base(*file)

	// Create a graphite client
	g, err := graphite.New(*graphiteAddress)
	if err != nil {
		log.Fatal(err)
	}
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
	go displayer.Display(g)

	// Alert
	var alerter Alerter
	alerter = CommonLogAlert{}
	go alerter.Alert(g, *threshold)

	// Process file
	for line := range t.Lines {
		err := logParser.LogParse(c, line.Text)
		if err != nil {
			log.Println(err)
		}
	}
}
