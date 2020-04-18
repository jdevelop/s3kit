package cmd

import "github.com/spf13/pflag"

const (
	allFlagName     = "all"
	latestFlagName  = "latest"
	versionFlagName = "version"
)

var accessConfig struct {
	version string
	latest  bool
	all     bool
}

func initConfig(f *pflag.FlagSet) {
	f.BoolVar(&accessConfig.latest, latestFlagName, true, "Apply to latest version of object(s)")
	f.BoolVar(&accessConfig.all, allFlagName, false, "Apply to all versions of object(s)")
	f.StringVar(&accessConfig.version, versionFlagName, "", "Apply to a specific version")
}
