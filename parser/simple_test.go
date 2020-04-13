package parser

import (
	"os"
	"testing"

	"github.com/jdevelop/s3kit/model"
	"github.com/stretchr/testify/require"
)

func TestSimpleParser(t *testing.T) {
	p := NewSimpleParser()
	f, err := os.Open("testdata/log.txt")
	require.NoError(t, err)
	defer f.Close()
	res := make([]model.S3AccessLogSimple, 0)
	require.NoError(t, p.ParseSimple(f, func(r *model.S3AccessLogSimple) bool {
		res = append(res, *r)
		return true
	}))
	require.Equal(t, 10, len(res))
}
