# httplog-analyzer

Analyze http logs tailing a file and showing metrics about the different parameters.

Current implementation only works with w3c-formatted HTTP access log (https://www.w3.org/Daemon/User/Config/Logging.html) , but is easy to add new logs format implementing the interface:

```go
// LogParser parse lines and sends stats to a Statsd server
type LogParser interface {
	LogParse(c *statsd.Client, line string) error
}
```

The corresponding code is in the file [logparse.go](logparse.go).

## Architecture

The monitoring system is based in 3 components:

* *httplog-analyzer*

1. parse the logs and sends the metrics to a statsd daemon.

2. show metrics in the console querying the data stored in InfluxDB.

3. alerts based on teh rate of request per seconds received.

* *telegraf*: StatsD collector that send the data to InfluxDB.

* *InfluxDB*: time series database that stores the data.


## Install

You can install the *httplog-analyzer* using `go build` or with the docker image provided.

You need to provide a StatsD collector and a storage, in this case we are going to use *telegraf* and *InfluxDB* 


## How to use it

The log analyzer is parametrizable with these options:

```sh
./httplog-analyzer -h
flag needs an argument: -h
Usage of ./httplog-analyzer:
  -f string
        log file (default "/tmp/access.log")
  -i string
        InfluxDB server address (default "http://localhost:8086")
  -h string
        help
  -s string
        Statsd server address (default "localhost:8125")
  -t int
        Threshold requests per second averaged over a 2 minutes slot (default 10)
```

If you don't have installed all the components, you can run a demo using docker.

1. Make sure you have installed [docker](https://docs.docker.com/install/) and [docker compose](https://docs.docker.com/compose/)

2. Run the compose file:

```sh
docker-compose up
```

3.



## UI

You can install Grafana to consume the data stored in InfluxDB, just configure it to [use InfluxDB in Grafana as data source](https://grafana.com/docs/grafana/latest/features/datasources/influxdb/)

## Extensions

*httplog-analyzer* can work with any StatsD daemon, per example, you can configure the client to connect to the Datadog agent or any other system that is able to consume StatsD data and see the data on those services.
