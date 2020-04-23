package cmd

import (
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
