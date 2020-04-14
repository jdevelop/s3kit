package cmd

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jdevelop/s3kit/model"
	"github.com/jdevelop/s3kit/parser"
	"github.com/spf13/cobra"
)

var objectsPerPage int64 = 100

type batch struct {
	bucket  string
	objects []*s3.Object
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "print S3 Access logs as JSON",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sess := session.New()
		svc := s3.New(sess)
		batchChan := make(chan batch)
		mChan := make(chan model.S3AccessLogSimple, 100)
		var wg, printer sync.WaitGroup
		wg.Add(globalOpts.workers)
		p := parser.NewSimpleParser()
		go func() {
			printer.Add(1)
			defer printer.Done()
			jsonner := json.NewEncoder(os.Stdout)
			for v := range mChan {
				jsonner.Encode(v)
			}
		}()
		for i := 0; i < globalOpts.workers; i++ {
			go func(svc *s3.S3) {
				defer wg.Done()
				for batch := range batchChan {
					for _, o := range batch.objects {
						res, err := svc.GetObject(&s3.GetObjectInput{
							Bucket: &batch.bucket,
							Key:    o.Key,
						})
						if err != nil {
							log.Printf("Error reading s3://%s/%s : %+v", batch.bucket, *o.Key, err)
							continue
						}
						p.ParseSimple(res.Body, func(m *model.S3AccessLogSimple) bool {
							if m.Time.Before(logsConfig.endDate.Time) && m.Time.After(logsConfig.startDate.Time) {
								mChan <- *m
							}
							return true
						})
						res.Body.Close()
					}
				}
			}(svc)
		}
		for _, url := range args {
			bucket, prefix, err := fromS3(url)
			if err != nil {
				return err
			}
			if err := svc.ListObjectsPages(&s3.ListObjectsInput{
				Bucket:  &bucket,
				Prefix:  &prefix,
				MaxKeys: &objectsPerPage,
			}, func(res *s3.ListObjectsOutput, last bool) bool {
				batchChan <- batch{
					bucket:  bucket,
					objects: res.Contents,
				}
				return true
			}); err != nil {
				return err
			}
		}
		close(batchChan)
		wg.Wait()
		close(mChan)
		printer.Wait()
		return nil
	},
}

var logsConfig = struct {
	startDate flagTime
	endDate   flagTime
}{
	endDate: flagTime{time.Now().AddDate(0, 0, 1).Truncate(time.Hour * 24)},
}

type flagTime struct {
	time.Time
}

func (t *flagTime) Set(value string) error {
	p, err := time.Parse("2006-01-02", value)
	if err != nil {
		return err
	}
	t.Time = p
	return nil
}

func (t *flagTime) String() string {
	return t.Time.String()
}

func (t *flagTime) Type() string {
	return "Date"
}

func init() {
	pf := logsCmd.PersistentFlags()
	pf.VarP(&logsConfig.startDate, "start", "s", "start date ( YYYY-MM-DD )")
	pf.VarP(&logsConfig.endDate, "end", "e", "end date ( YYYY-MM-DD )")
	rootCmd.AddCommand(logsCmd)
}
