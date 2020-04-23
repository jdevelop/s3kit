package cmd

import (
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"
)

type accessFuncT func(bucket string, o *s3.ObjectVersion) error

func accessFuncBuilder(op accessFuncT) accessFuncT {
	switch {
	case accessConfig.all:
		return func(bucket string, o *s3.ObjectVersion) error {
			return op(bucket, o)
		}
	case accessConfig.version != "":
		return func(bucket string, o *s3.ObjectVersion) error {
			if o.VersionId != nil && *o.VersionId == accessConfig.version {
				return op(bucket, o)
			}
			return nil
		}
	default: // use latest
		return func(bucket string, o *s3.ObjectVersion) error {
			if o.IsLatest != nil && *o.IsLatest {
				return op(bucket, o)
			}
			return nil
		}
	}
}

func run(urls []string, holdFunc accessFuncT) error {
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

	for _, url := range urls {
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
