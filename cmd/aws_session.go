package cmd

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	once sync.Once
	svc  *s3.S3
	sess *session.Session
)

func _init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = s3.New(sess)
}

func getS3() *s3.S3 {
	once.Do(_init)
	return svc
}

func getSession() *session.Session {
	once.Do(_init)
	return sess
}
