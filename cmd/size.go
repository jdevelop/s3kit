package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type pathSpec struct {
	bucket string
	prefix string
}
type SizeSpec struct {
	Path  string
	Count uint64
	Size  uint64
}

var sizeCmd = &cobra.Command{
	Use:   "size",
	Short: "calculate size of S3 location",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sess := session.Must(session.NewSession())
		svc := s3.New(sess)

		specsChan := make(chan pathSpec, 100)
		sizesChan := make(chan SizeSpec, 100)

		sizes := make([]SizeSpec, 0, len(args))

		var wg, sg sync.WaitGroup
		sg.Add(1)
		wg.Add(globalOpts.workers)

		go func() {
			defer sg.Done()
			for sizeSpec := range sizesChan {
				sizes = append(sizes, sizeSpec)
			}
		}()

		for i := 0; i < globalOpts.workers; i++ {
			go func() {
				defer wg.Done()
				for spec := range specsChan {
					var ss SizeSpec
					ss.Path = fmt.Sprintf("s3://%s/%s", spec.bucket, spec.prefix)
					if err := svc.ListObjectsPages(&s3.ListObjectsInput{
						Bucket: &spec.bucket,
						Prefix: &spec.prefix,
					}, func(res *s3.ListObjectsOutput, last bool) bool {
						for _, o := range res.Contents {
							ss.Size += uint64(*o.Size)
						}
						ss.Count += uint64(len(res.Contents))
						return true
					}); err != nil {
						log.Fatalf("can't list objects at s3://%s/%s => %v", spec.bucket, spec.prefix, err)
					}
					sizesChan <- ss
				}
			}()
		}

		for _, url := range args {
			bucket, prefix, err := fromS3(url)
			if err != nil {
				return err
			}
			if sizeOpts.group {
				if err := svc.ListObjectsPages(&s3.ListObjectsInput{
					Delimiter: aws.String("/"),
					Bucket:    &bucket,
					Prefix:    &prefix,
				}, func(res *s3.ListObjectsOutput, last bool) bool {
					for _, pfx := range res.CommonPrefixes {
						specsChan <- pathSpec{
							bucket: bucket,
							prefix: *pfx.Prefix,
						}
					}
					return true
				}); err != nil {
					log.Fatalf("can't list objects at s3://%s/%s => %v", bucket, prefix, err)
				}
			} else {
				specsChan <- pathSpec{
					bucket: bucket,
					prefix: prefix,
				}
			}
		}

		close(specsChan)
		wg.Wait()
		close(sizesChan)
		sg.Wait()
		switch {
		case sizeOpts.asJson:
			return json.NewEncoder(os.Stdout).Encode(sizes)
		default:
			var hmnz func(uint64) string
			if !sizeOpts.raw {
				hmnz = humanize.Bytes
			} else {
				hmnz = func(x uint64) string {
					return strconv.FormatUint(x, 10)
				}
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Path", "Count", "Size"})
			var totalSize, totalCount uint64 = 0, 0
			for _, size := range sizes {
				totalSize += size.Size
				totalCount += size.Count
				table.Append([]string{size.Path, strconv.FormatUint(size.Count, 10), hmnz(size.Size)})
			}
			table.SetFooter([]string{"Total:", strconv.FormatUint(totalCount, 10), hmnz(totalSize)})
			table.Render()
		}
		return nil
	},
}

func init() {
	pf := sizeCmd.Flags()
	pf.BoolVarP(&sizeOpts.group, "group", "g", false, "group sizes by top-level folders")
	pf.BoolVar(&sizeOpts.asJson, "json", false, "output as JSON array")
	pf.BoolVar(&sizeOpts.raw, "raw", false, "raw numbers, no human-formatted size")
	rootCmd.AddCommand(sizeCmd)
}

var sizeOpts struct {
	group  bool
	asJson bool
	raw    bool
}
