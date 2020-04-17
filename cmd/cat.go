package cmd

import (
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
)

var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "cat S3 file(s)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := getS3()
		for _, url := range args {
			bucket, prefix, err := fromS3(url)
			if err != nil {
				return err
			}
			if err := svc.ListObjectsPages(&s3.ListObjectsInput{
				Bucket: &bucket,
				Prefix: &prefix,
			}, func(res *s3.ListObjectsOutput, last bool) bool {
				for _, o := range res.Contents {
					val, err := svc.GetObject(&s3.GetObjectInput{
						Bucket: &bucket,
						Key:    o.Key,
					})
					if err != nil {
						return false
					}
					if err := func() error {
						defer val.Body.Close()
						var reader io.Reader
						switch {
						case strings.HasSuffix(*o.Key, ".gz") || strings.HasSuffix(*o.Key, ".gzip"):
							if r, err := gzip.NewReader(val.Body); err != nil {
								reader = val.Body
							} else {
								reader = r
							}
						case strings.HasSuffix(*o.Key, ".bz2"):
							if r := bzip2.NewReader(val.Body); err != nil {
								reader = val.Body
							} else {
								reader = r
							}
						default:
							reader = val.Body
						}
						_, err := io.Copy(os.Stdout, reader)
						return err
					}(); err != nil {
						panic(err)
					}
				}
				return true
			}); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(catCmd)
}
