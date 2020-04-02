package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
)

// LogParser parse lines and sends stats to a Statsd server
type LogParser interface {
	LogParse(c *statsd.Client, line string) error
}

// CommonLog implements the LogParser interface
type CommonLog struct{}

// Common Logfile Format
// https://www.w3.org/Daemon/User/Config/Logging.html

// The common logfile format is as follows:
//    remotehost rfc931 authuser [date] "request" status bytes

// remotehost: Remote hostname (or IP number if DNS hostname is not available, or if DNSLookup is Off.
// rfc931: The remote logname of the user.
// authuser: The username as which the user has authenticated himself.
// [date]: Date and time of the request.
// "request": The request line exactly as it came from the client.
// status: The HTTP status code returned to the client.
// bytes: The content-length of the document transferred.

// 127.0.0.1 - mary [09/May/2018:16:00:42 +0000] "POST /api/user HTTP/1.0" 503 12
// ParsedLineCLF contains the CLF fields
type ParsedLineCLF struct {
	remotehost string
	rfc931     string
	authuser   string
	date       string
	request    string
	status     string
	bytes      int64
}

// LogParse parse the logs of Common Log Format
func (clog CommonLog) LogParse(c *statsd.Client, line string) error {
	l, err := clog.parse(line)
	if err != nil {
		return err
	}
	clog.send(c, l)
	return nil
}

// Send sends stats from the parsed line
func (clog CommonLog) send(c *statsd.Client, p ParsedLineCLF) error {
	// TODO: process errors, aggregate all
	c.Incr("requests.count", nil, 1)
	c.Incr(fmt.Sprintf("host.%s.count", p.remotehost), nil, 1)
	c.Incr(fmt.Sprintf("user.%s.count", p.authuser), nil, 1)
	c.Incr(fmt.Sprintf("status.%s.count", p.status), nil, 1)
	c.Count("bytes.count", p.bytes, nil, 1)
	return nil
}

// Parse the log file
func (clog CommonLog) parse(line string) (ParsedLineCLF, error) {
	var parsedLine ParsedLineCLF
	var err error
	fields := getFieldsFromLog(line)
	if len(fields) != 7 {
		return ParsedLineCLF{}, fmt.Errorf("Expected 7 fields, Received %d fields: %v", len(fields), fields)
	}

	// remotehost
	parsedLine.remotehost = fields[0]
	// rfc931
	parsedLine.rfc931 = fields[1]
	// authuser
	parsedLine.authuser = fields[2]
	// date
	parsedLine.date = fields[3]
	// layout := "02/Jan/2006:15:04:05 -0700"
	// t, err := time.Parse(layout, date)
	// if err != nil {
	//	return ParsedLine{}, fmt.Errorf("Error parsing, expected an date and received %v: %v", date, err)
	//}
	// TODO: Compare with current time and Warn or Error that the data is old
	// request
	parsedLine.request = fields[4]
	// status
	parsedLine.status = fields[5]
	// bytes
	parsedLine.bytes, err = strconv.ParseInt(fields[6], 10, 64)
	if err != nil {
		return ParsedLineCLF{}, fmt.Errorf("Error parsing, expected an integer and received %v: %v", fields[6], err)
	}
	return parsedLine, nil
}

// GetFieldsFromLog parses an Apache CommonLog string and returns it split in fields
// since golang regexp doesn't allow to capture records and we know the log format beforehand
// we convert [ and ] to " split the line by " to obtain the fields inside brackets and quotes
// then we can just split by spaces the rest of the fields
func getFieldsFromLog(line string) []string {
	var fields []string
	// we need at least 7 fields
	if len(strings.Fields(line)) < 7 {
		return fields
	}
	// Replace [ and ] by "
	s := strings.Replace(line, "[", "\"", 1)
	s = strings.Replace(s, "]", "\"", 1)
	// Split by " so t[1] is the date and [3] the request
	t := strings.Split(s, "\"")
	fields = append(fields, strings.Fields(t[0])...)
	fields = append(fields, t[1], t[3])
	fields = append(fields, strings.Fields(t[4])...)
	return fields
}
