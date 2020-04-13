package parser

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/jdevelop/s3kit/model"
)

type simpleParser struct{}

// BucketOwner Bucket [Time] RemoteIP Requester RequestID Operation Key "RequestURI" HTTPstatus ErrorCode/- BytesSent/- ObjectSize TotalTime TurnAroundTime "Referer" "UserAgent" VersionId HostId SignatureVersion CipherSuite AuthenticationType HostHeader TLSversion

var (
	p       simpleParser
	matcher = regexp.MustCompile(`^(?P<BucketOwner>\S+?)\s+(?P<Bucket>\S+?)\s+\[(?P<Time>[^\]]+?)\]\s+(?P<RemoteIP>\S+?)\s+(?P<Requester>\S+?)\s+(?P<RequestID>\S+?)\s+(?P<Operation>\S+?)\s+(?P<Key>\S+?)\s+"(?P<RequestURI>[^"]+?)"\s+(?P<HTTPStatus>\S+?)\s+(?P<ErrorCode>\S+?)\s+(?P<BytesSent>\S+?)\s+(?P<ObjectSize>\S+?)\s+(?P<TotalTime>\S+?)\s+(?P<TurnaroundTime>\S+?)\s+"(?P<Referer>[^"]+?)"\s+"(?P<UserAgent>[^"]+?)"\s+(?P<VersionId>\S+?)\s+(?P<HostId>\S+?)\s+(?P<SignatureVersion>\S+?)\s+(?P<CipherSuite>\S+?)\s+(?P<AuthenticationType>\S+?)\s+(?P<HostHeader>\S+?)\s+(?P<TLSversion>\S+?)`)
	names   = matcher.SubexpNames()
)

func NewSimpleParser() *simpleParser {
	return &p
}

func (sp *simpleParser) ParseSimple(src io.Reader, visitor func(*model.S3AccessLogSimple) bool) error {
	var (
		lines = bufio.NewScanner(src)
		m     model.S3AccessLogSimple
	)
breakLine:
	for lines.Scan() {
		data := matcher.FindStringSubmatch(lines.Text())
		if len(data) != len(names) {
			fmt.Println(len(data), len(names))
			continue breakLine
		}
	lineMatch:
		for i, v := range data {
			switch names[i] {
			case "Time":
				d, err := time.Parse("02/Jan/2006:15:04:05 -0700", v) // 10/Apr/2020:22:03:06 +0000
				if err != nil {
					fmt.Printf("%#v\n", err)
					continue breakLine
				}
				m.Time = d
			case "RemoteIP":
				m.RemoteIP = net.ParseIP(v)
			case "Operation":
				m.Operation = v
			case "Key":
				m.Key = v
			case "RequestURI":
				m.RequestURI = v
			case "HTTPStatus":
				code, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					fmt.Printf("HTTPStatus: %#v\n", err)
					continue breakLine
				}
				m.HTTPStatus = uint(code)
			case "BytesSent":
				if v == "-" {
					continue
				}
				sent, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					fmt.Printf("BytesSent: %#v\n", err)
					continue breakLine
				}
				m.BytesSent = uint(sent)
			case "ObjectSize":
				if v == "-" {
					continue
				}
				size, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					fmt.Printf("ObjectSize: %#v\n", err)
					continue breakLine
				}
				m.ObjectSize = uint(size)
			case "TotalTime":
				if v == "-" {
					continue
				}
				totalTime, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					fmt.Printf("TotalTime: %#v\n", err)
					continue breakLine
				}
				m.FlightTime = time.Duration(totalTime) * time.Millisecond
			case "TurnaroundTime":
				if v == "-" {
					continue
				}
				turnaroundTime, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					fmt.Printf("TurnaroundTime: %#v\n", err)
					continue breakLine
				}
				m.TurnaroundTime = time.Duration(turnaroundTime) * time.Millisecond
			case "Referer":
				m.Referer = v
			case "UserAgent":
				m.UserAgent = v
				break lineMatch
			}
		}
		if !visitor(&m) {
			return nil
		}
	}
	return nil
}
