package parser

import (
	"io"

	"github.com/jdevelop/s3kit/model"
)

type S3AccessLogVisitor = func(*model.S3AccessLog) bool
type S3AccessLogVisitorSimple = func(*model.S3AccessLogSimple) bool

type S3LogParser interface {
	ParseFull(io.Reader, S3AccessLogVisitor) error
}

type S3SimpleLogParser interface {
	ParseSimple(io.Reader, S3AccessLogVisitorSimple) error
}
