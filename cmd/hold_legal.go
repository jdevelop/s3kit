package cmd

import (
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
)

var legalAdd = &cobra.Command{
	Use:          "add s3://bucket/key1 s3://bucket/prefix/ ...",
	Short:        "Add legal lock for given object(s)",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLegalAction(cmd, args, holdFuncBuilder(getS3(), s3.ObjectLockLegalHoldStatusOn))
	},
}

var legalRm = &cobra.Command{
	Use:          "rm s3://bucket/key1 s3://bucket/prefix/ ...",
	Short:        "Remove legal lock for given object(s)",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLegalAction(cmd, args, holdFuncBuilder(getS3(), s3.ObjectLockLegalHoldStatusOff))
	},
}

func init() {
	holdCmd.AddCommand(legalAdd, legalRm)
}

type holdFuncT func(bucket string, o *s3.ObjectVersion) error

func holdFuncBuilder(svc *s3.S3, opCode string) holdFuncT {
	putHold := func(bucket, key, version string) error {
		log.Infof("hold %s: s3://%s/%s@%s", opCode, bucket, key, version)
		_, err := svc.PutObjectLegalHold(
			&s3.PutObjectLegalHoldInput{
				Bucket: &bucket,
				Key:    &key,
				LegalHold: &s3.ObjectLockLegalHold{
					Status: &opCode,
				},
				VersionId: &version,
			},
		)
		return err
	}
	switch {
	case holdConfig.all:
		return func(bucket string, o *s3.ObjectVersion) error {
			return putHold(bucket, *o.Key, *o.VersionId)
		}
	case holdConfig.version != "":
		return func(bucket string, o *s3.ObjectVersion) error {
			if o.VersionId != nil && *o.VersionId == holdConfig.version {
				return putHold(bucket, *o.Key, holdConfig.version)
			}
			return nil
		}
	default: // use latest
		return func(bucket string, o *s3.ObjectVersion) error {
			if o.IsLatest != nil && *o.IsLatest {
				return putHold(bucket, *o.Key, *o.VersionId)
			}
			return nil
		}
	}
}

func runLegalAction(cmd *cobra.Command, args []string, holdFunc holdFuncT) error {
	svc := getS3()

	type Batch struct {
		bucket  string
		objects []*s3.ObjectVersion
	}

	var (
		batchChan = make(chan Batch)
		wg        sync.WaitGroup
	)

	log.Debugf("Sarting %d workers", globalOpts.workers)
	wg.Add(globalOpts.workers)

	for i := 0; i < globalOpts.workers; i++ {
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				log.Debugf("New batch: %+v", batch)
				for _, o := range batch.objects {
					log.Debugf("Processing s3://%s/%s", batch.bucket, *o.Key)
					if err := holdFunc(batch.bucket, o); err != nil {
						log.Fatalf("Can't process s3://%s/%s : %+v", batch.bucket, *o.Key, err)
					}
				}
			}
			log.Debug("Done")
		}()
	}

	for _, url := range args {
		bucket, prefix, err := fromS3(url)
		if err != nil {
			return err
		}
		if err := svc.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
			Bucket: &bucket,
			Prefix: &prefix,
		}, func(res *s3.ListObjectVersionsOutput, last bool) bool {
			batchChan <- Batch{
				bucket:  bucket,
				objects: res.Versions,
			}
			log.Debug("Sending batch")
			return true
		}); err != nil {
			log.Errorf("can't list objects at %s: %v", url, err)
			close(batchChan)
			return err
		}
		close(batchChan)
	}
	log.Debug("Done")
	wg.Wait()
	log.Debug("Complete")
	return nil
}
