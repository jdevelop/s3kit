package cmd

import (
	"errors"
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
)

const (
	allFlagName     = "all"
	latestFlagName  = "latest"
	versionFlagName = "version"
)

var holdOptsError = errors.New("Must specify either --" + allFlagName + " or --" + latestFlagName + " or --" + versionFlagName + " <version>")

var holdCmd = &cobra.Command{
	Use:     "hold",
	Short:   "Add/Remove legal lock for given object(s)",
	Aliases: []string{"unhold"},
	Args:    cobra.MinimumNArgs(1),
	PreRunE: func(*cobra.Command, []string) error {
		if holdConfig.all || holdConfig.latest || holdConfig.version != "" {
			return nil
		}
		return holdOptsError
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := s3.New(session.Must(session.NewSession()))

		type Batch struct {
			bucket  string
			objects []*s3.ObjectVersion
		}
		type holdFuncT func(bucket string, o *s3.ObjectVersion) error

		var (
			batchChan       = make(chan Batch)
			wg              sync.WaitGroup
			holdFunc        holdFuncT
			holdFuncBuilder = func(opCode string) holdFuncT {
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
				case holdConfig.latest:
					return func(bucket string, o *s3.ObjectVersion) error {
						if o.IsLatest != nil && *o.IsLatest {
							return putHold(bucket, *o.Key, *o.VersionId)
						}
						return nil
					}
				case holdConfig.all:
					return func(bucket string, o *s3.ObjectVersion) error {
						return putHold(bucket, *o.Key, *o.VersionId)
					}
				default:
					return func(bucket string, o *s3.ObjectVersion) error {
						if o.VersionId != nil && *o.VersionId == holdConfig.version {
							return putHold(bucket, *o.Key, holdConfig.version)
						}
						return nil
					}
				}
			}
		)

		switch cmd.CalledAs() {
		case "hold":
			holdFunc = holdFuncBuilder(s3.ObjectLockLegalHoldStatusOn)
		case "unhold":
			holdFunc = holdFuncBuilder(s3.ObjectLockLegalHoldStatusOff)
		}

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
	},
}

func init() {
	f := holdCmd.Flags()
	f.BoolVar(&holdConfig.latest, latestFlagName, false, "Apply to latest version of object(s)")
	f.BoolVar(&holdConfig.all, allFlagName, false, "Apply to all versions of object(s)")
	f.StringVar(&holdConfig.version, versionFlagName, "", "Apply to a specific version")
	rootCmd.AddCommand(holdCmd)
}

func anyErr(err ...error) error {
	for _, e := range err {
		if e != nil {
			return e
		}
	}
	return nil
}

var holdConfig struct {
	version string
	latest  bool
	all     bool
}
