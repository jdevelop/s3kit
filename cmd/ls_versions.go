package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type Version struct {
	VersionId    string    `json:"version_id" yaml:"version_id"`
	LastModified time.Time `json:"last_modified" yaml:"last_modified"`
	Latest       bool      `json:"latest" yaml:"latest"`
}

type PathVersion struct {
	Path     string    `json:"path" yaml:"path"`
	Versions []Version `json:"versions" yaml:"versions"`
	Latest   string    `json:"latest" yaml:"latest"`
}

func mapValues(src map[string]*PathVersion) []*PathVersion {
	v := make([]*PathVersion, len(src))
	i := 0
	for _, val := range src {
		v[i] = val
		i++
	}
	return v
}

var lsVersions = &cobra.Command{
	Use:   "versions s3://bucket/folder/ s3://bucket/folder/prefix ...",
	Short: "List object version(s)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, urls []string) error {
		svc := getS3()
		var renderF func(string, map[string]*PathVersion) error
		switch {
		case lsConfig.asJson:
			renderF = func(bucket string, keymap map[string]*PathVersion) error {
				return json.NewEncoder(os.Stdout).Encode(mapValues(keymap))
			}
		case lsConfig.asYaml:
			renderF = func(bucket string, keymap map[string]*PathVersion) error {
				return yaml.NewEncoder(os.Stdout).Encode(mapValues(keymap))
			}
		default:
			renderF = func(bucket string, keymap map[string]*PathVersion) error {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Path", "Version", "Last Modified", "Latest"})
				table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})
				var verString, timeString, latestString strings.Builder
				for _, row := range keymap {
					verString.WriteString(row.Versions[0].VersionId)
					timeString.WriteString(row.Versions[0].LastModified.Format("2006-01-02 15:04:05 -0700 MST"))
					if row.Versions[0].Latest {
						latestString.WriteString("*")
					} else {
						latestString.WriteString(" ")
					}
					for _, v := range row.Versions[1:] {
						verString.WriteString("\n")
						verString.WriteString(v.VersionId)
						timeString.WriteString("\n")
						timeString.WriteString(v.LastModified.Format("2006-01-02 15:04:05 -0700 MST"))
						latestString.WriteString("\n")
						if v.Latest {
							latestString.WriteString("*")
						} else {
							latestString.WriteString(" ")
						}
					}
					table.Append([]string{row.Path, verString.String(), timeString.String(), latestString.String()})
					verString.Reset()
				}
				table.Render()
				return nil
			}
		}
		for _, url := range urls {
			bucket, prefix, err := fromS3(url)
			if err != nil {
				return err
			}
			keysMap := make(map[string]*PathVersion)
			if err := svc.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
				Bucket: &bucket,
				Prefix: &prefix,
			}, func(res *s3.ListObjectVersionsOutput, last bool) bool {
				for _, ver := range res.Versions {
					if v, ok := keysMap[*ver.Key]; ok {
						v.Versions = append(v.Versions, Version{
							VersionId:    *ver.VersionId,
							LastModified: *ver.LastModified,
							Latest:       *ver.IsLatest,
						})
						if *ver.IsLatest {
							v.Latest = *ver.VersionId
						}
					} else {
						v = &PathVersion{
							Path:     fmt.Sprintf("s3://%s/%s", bucket, *ver.Key),
							Versions: []Version{{VersionId: *ver.VersionId, LastModified: *ver.LastModified, Latest: *ver.IsLatest}},
						}
						if *ver.IsLatest {
							v.Latest = *ver.VersionId
						}
						keysMap[*ver.Key] = v
					}
				}
				return true
			}); err != nil {
				return err
			}
			renderF(bucket, keysMap)
		}
		return nil
	},
}

func init() {
	lsCmd.AddCommand(lsVersions)
}
