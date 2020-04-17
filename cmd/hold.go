package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

const (
	allFlagName     = "all"
	latestFlagName  = "latest"
	versionFlagName = "version"
)

var holdOptsError = errors.New("Must specify either --" + allFlagName + " or --" + latestFlagName + " or --" + versionFlagName + " <version>")

var holdCmd = &cobra.Command{
	Use:          "hold",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
}

func init() {
	f := holdCmd.PersistentFlags()
	f.BoolVar(&holdConfig.latest, latestFlagName, true, "Apply to latest version of object(s)")
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
