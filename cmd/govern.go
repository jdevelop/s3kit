package cmd

import (
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
)

var governCmd = &cobra.Command{
	Use:          "governance",
	Short:        "Add/remove governance lock",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
}

var governAdd = &cobra.Command{
	Use:          "add s3://bucket/key1 s3://bucket/prefix/ ...",
	Short:        "Add governance lock for given object(s)",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(args, accessFuncBuilder(governOp(getS3(), "ON")))
	},
}

var governRm = &cobra.Command{
	Use:          "rm s3://bucket/key1 s3://bucket/prefix/ ...",
	Short:        "Remove governance lock for given object(s)",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(args, accessFuncBuilder(governOp(getS3(), "OFF")))
	},
}

func init() {
	f := governCmd.PersistentFlags()
	initVersionsConfig(f)
	governAdd.Flags().DurationVar(&govConf.duration, "expire", 0, "governance lock duration (1m, 1h etc)")
	governAdd.MarkFlagRequired("expire")
	governCmd.AddCommand(governAdd, governRm)
	lockRoot.AddCommand(governCmd)
}

var (
	govMode = s3.ObjectLockRetentionModeGovernance
	bypass  = true
)

func governOp(svc *s3.S3, opCode string) accessFuncT {
	switch opCode {
	case "ON":
		return func(bucket string, o *s3.ObjectVersion) error {
			log.Infof("governance %s: s3://%s/%s@%s", opCode, bucket, *o.Key, *o.VersionId)
			expireAt := time.Now().UTC().Add(govConf.duration)
			_, err := svc.PutObjectRetention(
				&s3.PutObjectRetentionInput{
					Bucket: &bucket,
					Key:    o.Key,
					Retention: &s3.ObjectLockRetention{
						Mode:            &govMode,
						RetainUntilDate: &expireAt,
					},
					VersionId: o.VersionId,
				},
			)
			return err
		}
	case "OFF":
		return func(bucket string, o *s3.ObjectVersion) error {
			log.Infof("governance %s: s3://%s/%s@%s", opCode, bucket, *o.Key, *o.VersionId)
			expireAt := time.Now().UTC().Add(1 * time.Second)
			_, err := svc.PutObjectRetention(
				&s3.PutObjectRetentionInput{
					Bucket: &bucket,
					Key:    o.Key,
					Retention: &s3.ObjectLockRetention{
						Mode:            &govMode,
						RetainUntilDate: &expireAt,
					},
					BypassGovernanceRetention: &bypass,
					VersionId:                 o.VersionId,
				},
			)
			return err
		}
	}
	panic(opCode + "not implemented")
}

var govConf struct {
	duration time.Duration
}
