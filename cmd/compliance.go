package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
)

var complCmd = &cobra.Command{
	Use:          "compliance s3://bucket/key1 s3://bucket/prefix/ ...",
	Short:        "Add compliance lock",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, urls []string) error {
		reader := bufio.NewReader(os.Stdin)
		return run(urls, accessFuncBuilder(complOp(getS3(), reader)))
	},
}

func init() {
	f := complCmd.Flags()
	initConfig(f)
	complCmd.Flags().DurationVar(&complianceConf.duration, "expire", 0, "compliance lock duration (1m, 1h etc)")
	complCmd.MarkFlagRequired("expire")
	lockRoot.AddCommand(complCmd)
}

var (
	complianceMode = s3.ObjectLockRetentionModeCompliance
)

func complOp(svc *s3.S3, rdr *bufio.Reader) opFunc {
	return func(bucket, key, version string) error {
		expireAt := time.Now().UTC().Add(complianceConf.duration)
		fmt.Printf("Locking s3://%s/%s version %s expires %s, proceed? (y/N):", bucket, key, version, expireAt.Format("2006-01-02 15:04:05"))
		answer, _, err := rdr.ReadLine()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case string(answer) == "Y" || string(answer) == "y":
			_, err := svc.PutObjectRetention(
				&s3.PutObjectRetentionInput{
					Bucket: &bucket,
					Key:    &key,
					Retention: &s3.ObjectLockRetention{
						Mode:            &complianceMode,
						RetainUntilDate: &expireAt,
					},
					VersionId: &version,
				},
			)
			return err
		default:
			return nil
		}
		return nil
	}
}

var complianceConf struct {
	duration time.Duration
}
