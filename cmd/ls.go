package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List versions and/legal holds and locks",
}

func init() {
	rootCmd.AddCommand(lsCmd)
	pf := lsCmd.PersistentFlags()
	pf.BoolVar(&lsConfig.asJson, "json", false, "JSON output")
	pf.BoolVar(&lsConfig.asYaml, "yaml", false, "YAML output")
	pf.BoolVar(&lsConfig.asTable, "table", true, "ASCII table output")
}

var lsConfig struct {
	asJson  bool
	asYaml  bool
	asTable bool
}

type Version struct {
	VersionId    string    `json:"version_id" yaml:"version_id"`
	LastModified time.Time `json:"last_modified" yaml:"last_modified"`
	Latest       bool      `json:"latest" yaml:"latest"`
}

type VersionTag struct {
	Version `json:"version" yaml:"version"`
	Tag     []string `json:"tag" yaml:"tag"`
}

type PathVersion struct {
	Path     string    `json:"path" yaml:"path"`
	Versions []Version `json:"versions" yaml:"versions"`
	Latest   string    `json:"latest" yaml:"latest"`
}

type PathVersionTag struct {
	Path     string       `json:"path" yaml:"path"`
	Versions []VersionTag `json:"versions" yaml:"versions"`
}

func mapPathVersions(src map[string]*PathVersion) []*PathVersion {
	v := make([]*PathVersion, len(src))
	i := 0
	for _, val := range src {
		v[i] = val
		i++
	}
	return v
}

func mapPathVersionTags(src map[string]*PathVersionTag) []*PathVersionTag {
	v := make([]*PathVersionTag, len(src))
	i := 0
	for _, val := range src {
		v[i] = val
		i++
	}
	return v
}
