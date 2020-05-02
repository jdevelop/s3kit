package cmd

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
)

var tagRoot = &cobra.Command{
	Use:   "tag",
	Short: "Tag S3 object(s)",
}

func getTagMap(expectValue bool) map[string]string {
	tagSet := make(map[string]string)
	for _, t := range tagFlags.tags {
		i := strings.IndexRune(t, '=')
		switch {
		case expectValue && i > 0:
			tagSet[t[:i]] = t[i+1:]
		case !expectValue && i > 0:
			tagSet[t[:i]] = ""
		case !expectValue:
			tagSet[t] = ""
		}
	}
	return tagSet
}

var tagAdd = &cobra.Command{
	Use:          "add s3://bucket/folder/ s3://bucket/folder/prefix ...",
	Short:        "Add tag(s) to S3 object(s)",
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, urls []string) error {
		svc := getS3()
		return run(urls, accessFuncBuilder(func(bucket string, o *s3.ObjectVersion) error {
			tagResp, err := svc.GetObjectTagging(&s3.GetObjectTaggingInput{
				Bucket:    &bucket,
				Key:       o.Key,
				VersionId: o.VersionId,
			})
			if err != nil {
				return err
			}
			ts := tagResp.TagSet
			newTags := getTagMap(true)
			for i := range ts {
				if v, ok := newTags[*ts[i].Key]; ok {
					ts[i].Value = aws.String(v) // copy
					delete(newTags, *ts[i].Key)
				}
			}
			for k, v := range newTags {
				ts = append(ts, &s3.Tag{
					Key:   aws.String(k),
					Value: aws.String(v),
				})
			}

			_, err = svc.PutObjectTagging(&s3.PutObjectTaggingInput{
				Bucket:    &bucket,
				Key:       o.Key,
				VersionId: o.VersionId,
				Tagging: &s3.Tagging{
					TagSet: ts,
				},
			})
			return err
		}))
	},
}

var tagRm = &cobra.Command{
	Use:          "rm s3://bucket/folder/ s3://bucket/folder/prefix ...",
	Short:        "remove tag(s) from S3 object(s)",
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, urls []string) error {
		svc := getS3()
		return run(urls, accessFuncBuilder(func(bucket string, o *s3.ObjectVersion) error {
			tagResp, err := svc.GetObjectTagging(&s3.GetObjectTaggingInput{
				Bucket:    &bucket,
				Key:       o.Key,
				VersionId: o.VersionId,
			})
			if err != nil {
				return err
			}
			ts := tagResp.TagSet
			rmTags := getTagMap(false)
			i := 0
			for i < len(ts) {
				if _, ok := rmTags[*ts[i].Key]; ok {
					// remove an element in an array by swapping it with last element and reducing array size
					ts[len(ts)-1], ts[i] = ts[i], ts[len(ts)-1]
					ts = ts[:len(ts)-1]
					continue
				}
				i++
			}
			_, err = svc.PutObjectTagging(&s3.PutObjectTaggingInput{
				Bucket:    &bucket,
				Key:       o.Key,
				VersionId: o.VersionId,
				Tagging: &s3.Tagging{
					TagSet: ts,
				},
			})
			return err
		}))
	},
}

func init() {
	tagAdd.Flags().StringSliceVar(&tagFlags.tags, "tags", nil, "tags as --tags 'tag1=value1,tag2=value2' or multiple --tags ... options")
	tagAdd.MarkFlagRequired("tags")
	tagRm.Flags().StringSliceVar(&tagFlags.tags, "tags", nil, "tags as --tags 'tag1,tag2' or multiple --tags ... options")
	tagRm.MarkFlagRequired("tags")
	initVersionsConfig(tagRoot.PersistentFlags())
	tagRoot.AddCommand(tagAdd, tagRm)
	rootCmd.AddCommand(tagRoot)
}

var tagFlags struct {
	tags []string
}
