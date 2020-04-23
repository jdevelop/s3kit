package cmd

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
)

var legalCmd = &cobra.Command{
	Use:          "legal",
	Short:        "Add/remove legal hold",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
}

var legalAdd = &cobra.Command{
	Use:          "add s3://bucket/key1 s3://bucket/prefix/ ...",
	Short:        "Add legal hold for given object(s)",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(args, accessFuncBuilder(holdOp(getS3(), s3.ObjectLockLegalHoldStatusOn)))
	},
}

var legalRm = &cobra.Command{
	Use:          "rm s3://bucket/key1 s3://bucket/prefix/ ...",
	Short:        "Remove legal hold for given object(s)",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(args, accessFuncBuilder(holdOp(getS3(), s3.ObjectLockLegalHoldStatusOff)))
	},
}

func init() {
	f := legalCmd.PersistentFlags()
	initVersionsConfig(f)
	lockRoot.AddCommand(legalCmd)
	legalCmd.AddCommand(legalAdd, legalRm)
}

func holdOp(svc *s3.S3, opCode string) accessFuncT {
	return func(bucket string, o *s3.ObjectVersion) error {
		log.Infof("hold %s: s3://%s/%s@%s", opCode, bucket, *o.Key, *o.VersionId)
		_, err := svc.PutObjectLegalHold(
			&s3.PutObjectLegalHoldInput{
				Bucket: &bucket,
				Key:    o.Key,
				LegalHold: &s3.ObjectLockLegalHold{
					Status: &opCode,
				},
				VersionId: o.VersionId,
			},
		)
		return err
	}
}
