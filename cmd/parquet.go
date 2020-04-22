package cmd

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	ps3 "github.com/xitongsys/parquet-go-source/s3"
	parquet "github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
)

var parquetCmd = &cobra.Command{
	Use:   "parquet",
	Short: "Parquet files explorer",
}
var parquetSchema = &cobra.Command{
	Use:   "schema s3://bucket/prefix/key s3://bucket/prefix/ ...",
	Short: "Print parquet files schema",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, urls []string) error {
		ps3.SetActiveSession(getSession())
		var printFunc func([]*parquet.SchemaElement) error
		if parquetConf.isJson {
			printFunc = printSchemaJson
		} else {
			printFunc = printSchemaTableSimple
		}
		for _, url := range urls {
			log.Debugf("processing %s", url)
			bucket, key, err := fromS3(url)
			if err != nil {
				return err
			}
			var processed int
			svc := getS3()
			if err := svc.ListObjectsPages(&s3.ListObjectsInput{
				Bucket: &bucket,
				Prefix: &key,
			}, func(res *s3.ListObjectsOutput, last bool) bool {
				for _, obj := range res.Contents {
					if strings.HasSuffix(*obj.Key, "_SUCCESS") || strings.HasSuffix(*obj.Key, ".crc") {
						continue
					}
					if processed >= parquetConf.maxKeys {
						return false
					}
					pf, err := ps3.NewS3FileReader(context.TODO(), bucket, *obj.Key)
					if err != nil {
						log.Errorf("can't create file s3://%s/%s : %+v", bucket, *obj.Key, err)
						return false
					}
					r, err := reader.NewParquetReader(pf, nil, 0)
					if err != nil {
						log.Errorf("can't create reader from s3://%s/%s : %+v", bucket, *obj.Key, err)
						return false
					}
					printFunc(r.Footer.Schema)
					processed += 1
				}
				return false
			}); err != nil {
				return err
			}
		}
		return nil
	},
}

func printSchemaJson(schema []*parquet.SchemaElement) error {
	encoder := json.NewEncoder(os.Stdout)
	return encoder.Encode(schema)
}

func printSchemaTableSimple(schema []*parquet.SchemaElement) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type"})
	for _, s := range schema {
		if s.Type != nil {
			table.Append([]string{s.Name, s.Type.String()})
		} else {
			table.Append([]string{s.Name, "<COMPOSITE>"})
		}
	}
	table.Render()
	return nil
}

func init() {
	parquetCmd.AddCommand(parquetSchema)
	rootCmd.AddCommand(parquetCmd)
	pff := parquetSchema.Flags()
	pff.IntVar(&parquetConf.maxKeys, "keys", 1, "max parquet files to process")
	pff.BoolVar(&parquetConf.isJson, "json", false, "JSON output")
}

var parquetConf struct {
	maxKeys int
	isJson  bool
}
