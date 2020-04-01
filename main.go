package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hpcloud/tail"
)

const defaultFile = "/tmp/access.log"

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

type ParsedLine struct {
	remotehost string
	rfc931     string
	authuser   string
	date       string
	request    string
	status     string
	bytes      int
}

func main() {
	// Configuration
	file := flag.String("f", defaultFile, "log file")
	flag.Parse()
	if len(os.Args) > 3 {
		flag.PrintDefaults()
		os.Exit(1)
	}
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

	// Process file
	for line := range t.Lines {
		// TODO: Count each line
		l, err := Parse(line.Text)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println(l)

	}
}

func Parse(line string) (ParsedLine, error) {
	var parsedLine ParsedLine
	var err error
	fields := GetStringFromLog(line)
	if len(fields) != 7 {
		return ParsedLine{}, fmt.Errorf("Expected 7 fields, Received %d fields: %v", len(fields), fields)
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
	// statues
	parsedLine.status = fields[5]
	// bytes
	parsedLine.bytes, err = strconv.Atoi(fields[6])
	if err != nil {
		return ParsedLine{}, fmt.Errorf("Error parsing, expected an integer and received %v: %v", fields[6], err)
	}
	return parsedLine, nil
}

// GetStringFromLog parses an Apache CommonLog string and returns it split in fields
// since golang regexp doesn't allow to capture records and we know the log format beforehand
// we convert [ and ] to " split the line by " to obtain the fields inside brackets and quotes
// then we can just split by spaces the rest of the fields
func GetStringFromLog(line string) []string {
	var fields []string
	// we need at least 7 fields
	if len(strings.Fields(line)) < 7 {
		return fields
	}
	s := strings.Replace(line, "[", "\"", 1)
	s = strings.Replace(s, "]", "\"", 1)
	t := strings.Split(s, "\"")
	fields = append(fields, strings.Fields(t[0])...)
	fields = append(fields, t[1], t[3])
	fields = append(fields, strings.Fields(t[4])...)
	return fields
}
