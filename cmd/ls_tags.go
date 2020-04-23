package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var lsTags = &cobra.Command{
	Use:          "tags s3://bucket/folder/ s3://bucket/folder/prefix ...",
	Short:        "List tags for object(s)",
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, urls []string) error {
		type VTuple struct {
			bucket string
			o      *s3.ObjectVersion
		}
		processChan := make(chan VTuple, 100)
		var wg sync.WaitGroup
		wg.Add(1)
		keysMap := make(map[string]*PathVersionTag)
		go func() {
			defer wg.Done()
			for t := range processChan {
				ver := t.o
				res, err := svc.GetObjectTagging(&s3.GetObjectTaggingInput{
					Bucket:    &t.bucket,
					Key:       t.o.Key,
					VersionId: t.o.VersionId,
				})
				var tags []string
				if err != nil {
					log.Errorf("can't get tags for s3://%s/%s version %s", t.bucket, *t.o.Key, *t.o.VersionId)
				} else {
					tags = make([]string, len(res.TagSet))
					for i, ts := range res.TagSet {
						tags[i] = *ts.Key + "=" + *ts.Value
					}
				}
				var (
					v  *PathVersionTag
					ok bool
				)
				if v, ok = keysMap[*ver.Key]; ok {
					v.Versions = append(v.Versions, VersionTag{
						Version: Version{
							VersionId:    *ver.VersionId,
							LastModified: *ver.LastModified,
							Latest:       *ver.IsLatest,
						},
						Tag: tags,
					})
				} else {
					v = &PathVersionTag{
						Path: fmt.Sprintf("s3://%s/%s", t.bucket, *ver.Key),
						Versions: []VersionTag{
							{
								Version: Version{
									VersionId:    *ver.VersionId,
									LastModified: *ver.LastModified,
									Latest:       *ver.IsLatest,
								},
								Tag: tags,
							},
						},
					}
					keysMap[*ver.Key] = v
				}
			}
		}()
		var tagLister accessFuncT = func(bucket string, o *s3.ObjectVersion) error {
			processChan <- VTuple{
				bucket: bucket,
				o:      o,
			}
			return nil
		}
		if err := run(urls, accessFuncBuilder(tagLister)); err != nil {
			close(processChan)
			return err
		}
		close(processChan)
		wg.Wait()
		var renderF func(map[string]*PathVersionTag) error
		switch {
		case lsConfig.asJson:
			renderF = func(keymap map[string]*PathVersionTag) error {
				return json.NewEncoder(os.Stdout).Encode(mapPathVersionTags(keymap))
			}
		case lsConfig.asYaml:
			renderF = func(keymap map[string]*PathVersionTag) error {
				return yaml.NewEncoder(os.Stdout).Encode(mapPathVersionTags(keymap))
			}
		default:
			renderF = func(keymap map[string]*PathVersionTag) error {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Path", "Version", "Tags"})
				table.SetAutoMergeCells(true)
				table.SetRowLine(true)
				table.SetAutoMergeCellsByColumnIndex([]int{0, 1})
				for _, row := range keymap {
					for _, v := range row.Versions {
						table.Append([]string{row.Path, v.VersionId, strings.Join(v.Tag, "\n")})
					}
				}
				table.Render()
				return nil
			}
		}
		renderF(keysMap)
		return nil
	},
}

func init() {
	f := lsTags.Flags()
	initVersionsConfig(f)
	lsCmd.AddCommand(lsTags)
}
