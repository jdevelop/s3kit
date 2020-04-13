package model

import (
	"net"
	"time"
)

type S3AccessLog struct {
	Owner            string    `json:"owner"`
	Bucket           string    `json:"bucket"`
	Time             time.Time `json:"time"`
	RemoteIP         net.IP    `json:"remote_ip"`
	Requester        string    `json:"requester"`
	RequestID        string    `json:"request_id"`
	Operation        string    `json:"operation"`
	Key              string    `json:"key"`
	URI              string    `json:"uri"`
	HTTPStatus       uint      `json:"http_status"`
	ErrorCode        uint      `json:"error_code"`
	BytesSent        uint      `json:"bytes_sent"`
	ObjectSize       uint      `json:"object_size"`
	FlightTime       uint      `json:"flight_time"`
	TurnaroundTime   uint      `json:"turnaround_time"`
	Referer          string    `json:"referer"`
	UserAgent        string    `json:"user_agent"`
	Version          string    `json:"version"`
	HostId           string    `json:"host_id"`
	SignatureVersion string    `json:"signature_version"`
	CipherSuite      string    `json:"cipher_suite"`
	AuthType         string    `json:"auth_type"`
	HostHeader       string    `json:"host_header"`
	TLSVersion       string    `json:"tls_version"`
}

type S3AccessLogSimple struct {
	Time           time.Time     `json:"time"`
	RemoteIP       net.IP        `json:"remote_ip"`
	Operation      string        `json:"operation"`
	Key            string        `json:"key"`
	RequestURI     string        `json:"request_uri"`
	HTTPStatus     uint          `json:"http_status"`
	BytesSent      uint          `json:"bytes_sent"`
	ObjectSize     uint          `json:"object_size"`
	FlightTime     time.Duration `json:"flight_time"`
	TurnaroundTime time.Duration `json:"turnaround_time"`
	Referer        string        `json:"referer"`
	UserAgent      string        `json:"user_agent"`
}
