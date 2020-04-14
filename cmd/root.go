package cmd

import (
	"fmt"
	"log"
	"net/url"
	"runtime"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "s3kit",
	Short: "AWS S3 command line toolkit",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func fromS3(link string) (string, string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", "", err
	}
	if u.Host == "" {
		return "", "", fmt.Errorf("no host defined for %s", link)
	}
	return u.Host, u.Path[1:], nil
}

var globalOpts = struct {
	workers int
}{
	workers: runtime.NumCPU(),
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.IntVarP(&globalOpts.workers, "workers", "w", runtime.NumCPU(), "number of concurrent threads")
}
