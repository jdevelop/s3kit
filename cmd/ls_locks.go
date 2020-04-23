package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var lsLocks = &cobra.Command{
	Use:   "locks s3://bucket/folder/ s3://bucket/folder/prefix ...",
	Short: "List various locks on S3 object(s) (legal hold, governance/compliance retention)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, urls []string) error {
		type VTuple struct {
			bucket string
			o      *s3.ObjectVersion
		}
		processChan := make(chan VTuple, 100)
		var wg sync.WaitGroup
		wg.Add(1)
		keysMap := make(map[string]*PathVersionLocks)
		go func() {
			defer wg.Done()
			for t := range processChan {
				ver := t.o
				var (
					legalHoldStatus bool
					complianceExp   *time.Time
					governanceExp   *time.Time
				)
				if holdResult, err := svc.GetObjectLegalHold(&s3.GetObjectLegalHoldInput{
					Bucket:    &t.bucket,
					Key:       t.o.Key,
					VersionId: t.o.VersionId,
				}); err != nil {
					switch errT := err.(type) {
					case awserr.Error:
						if errT.Code() != "NoSuchObjectLockConfiguration" {
							log.Errorf("can't get legal hold for s3://%s/%s version %s: %+v", t.bucket, *t.o.Key, *t.o.VersionId, err)
						}
					default:
						log.Errorf("can't get legal hold for s3://%s/%s version %s: %+v", t.bucket, *t.o.Key, *t.o.VersionId, err)
					}
				} else if holdResult.LegalHold != nil {
					legalHoldStatus = *holdResult.LegalHold.Status == "ON"
				}
				if retentionRes, err := svc.GetObjectRetention(&s3.GetObjectRetentionInput{
					Bucket:    &t.bucket,
					Key:       t.o.Key,
					VersionId: t.o.VersionId,
				}); err != nil {
					switch errT := err.(type) {
					case awserr.Error:
						if errT.Code() != "NoSuchObjectLockConfiguration" {
							log.Errorf("can't get legal hold for s3://%s/%s version %s: %+v", t.bucket, *t.o.Key, *t.o.VersionId, err)
						}
					default:
						log.Errorf("can't get legal hold for s3://%s/%s version %s: %+v", t.bucket, *t.o.Key, *t.o.VersionId, err)
					}
				} else if retentionRes.Retention != nil {
					switch *retentionRes.Retention.Mode {
					case s3.ObjectLockRetentionModeCompliance:
						complianceExp = retentionRes.Retention.RetainUntilDate
					case s3.ObjectLockModeGovernance:
						governanceExp = retentionRes.Retention.RetainUntilDate
					}
				}
				var (
					v  *PathVersionLocks
					ok bool
				)
				vLock := VersionLocks{
					Version: Version{
						VersionId:    *ver.VersionId,
						LastModified: *ver.LastModified,
						Latest:       *ver.IsLatest,
					},
					LegalHold:           legalHoldStatus,
					ComplianceRetention: complianceExp,
					GovernanceRetention: governanceExp,
				}
				if v, ok = keysMap[*ver.Key]; ok {
					v.Versions = append(v.Versions, vLock)
				} else {
					v = &PathVersionLocks{
						Path:     fmt.Sprintf("s3://%s/%s", t.bucket, *ver.Key),
						Versions: []VersionLocks{vLock},
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
		var renderF func(map[string]*PathVersionLocks) error
		switch {
		case lsConfig.asJson:
			renderF = func(keymap map[string]*PathVersionLocks) error {
				return json.NewEncoder(os.Stdout).Encode(mapPathVersionLocks(keymap))
			}
		case lsConfig.asYaml:
			renderF = func(keymap map[string]*PathVersionLocks) error {
				return yaml.NewEncoder(os.Stdout).Encode(mapPathVersionLocks(keymap))
			}
		default:
			renderF = func(keymap map[string]*PathVersionLocks) error {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Path", "Version", "Hold", "Governance", "Compliance"})
				table.SetAutoMergeCells(true)
				table.SetRowLine(true)
				table.SetAutoMergeCellsByColumnIndex([]int{0, 1})
				for _, row := range keymap {
					for _, v := range row.Versions {
						tableRow := []string{row.Path, v.VersionId, "", "", ""}
						if v.LegalHold {
							tableRow[2] = "ON"
						}
						if v.GovernanceRetention != nil {
							tableRow[3] = v.GovernanceRetention.Format("2006-01-02 15:04:05 -0700 MST")
						}
						if v.ComplianceRetention != nil {
							tableRow[4] = v.ComplianceRetention.Format("2006-01-02 15:04:05 -0700 MST")
						}
						table.Append(tableRow)
					}
				}
				table.Render()
				return nil
			}
		}
		renderF(keysMap)
		return nil
		return nil
	},
}

func init() {
	initVersionsConfig(lsLocks.Flags())
	lsCmd.AddCommand(lsLocks)
}
