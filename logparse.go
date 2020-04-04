package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
)

// LogParser parse lines and sends stats to a Statsd server
type LogParser interface {
	LogParse(line string) error
}

// CommonLog implements the LogParser interface
type CommonLog struct {
	client   *statsd.Client
	filename string
}

// NewCommonLog returns a new CommonLog parses
func NewCommonLog(c *statsd.Client, f string) *CommonLog {
	return &CommonLog{
		client:   c,
		filename: f,
	}
}

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
func (c CommonLog) LogParse(line string) error {
	l, err := c.parse(line)
	if err != nil {
		return err
	}
	c.send(l)
	return nil
}

// Send sends stats from the parsed line
func (c CommonLog) send(p ParsedLineCLF) error {

	// TODO: process errors, aggregate all
	tag := "file" + ":" + c.filename
	// Obtain section
	section := getSectionFromRequest(p.request)
	tags := []string{tag, "user:" + p.authuser, "host:" + p.remotehost, "section:" + section, "status:" + p.status}

	c.client.Incr("requests.total", []string{tag}, 1)
	c.client.Incr("requests.remotehost.count", tags, 1)
	c.client.Incr("requests.user.count", tags, 1)
	c.client.Incr("requests.status.count", tags, 1)
	c.client.Incr("requests.section.count", tags, 1)
	// We can obtain bytes per user, host or section visited
	c.client.Count("requests.bytes.count", p.bytes, tags, 1)
	return nil
}

// Parse the log file
func (c CommonLog) parse(line string) (ParsedLineCLF, error) {
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

// TODO: we can get more data from the request
// "{method} {request} {protocol}"
func getSectionFromRequest(req string) string {
	s := strings.Fields(req)
	m := strings.SplitAfterN(s[1], "/", 3)
	return strings.TrimSuffix(m[1], "/")
}
