package cmd

import (
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

var rootCmd = &cobra.Command{
	Use:   "s3kit",
	Short: "AWS S3 command line toolkit",
	PersistentPreRunE: func(*cobra.Command, []string) error {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeTime = func(time.Time, zapcore.PrimitiveArrayEncoder) {}
		cfg.EncoderConfig.EncodeCaller = func(zapcore.EntryCaller, zapcore.PrimitiveArrayEncoder) {}
		switch {
		case globalOpts.debug:
			cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		case globalOpts.quiet:
			cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		default:
			cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		}
		l, err := cfg.Build()
		if err != nil {
			return err
		}
		log = l.Sugar()
		return nil
	},
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
	debug   bool
	quiet   bool
}{
	workers: runtime.NumCPU(),
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.IntVarP(&globalOpts.workers, "workers", "w", runtime.NumCPU(), "number of concurrent threads")
	f := rootCmd.Flags()
	f.BoolVar(&globalOpts.debug, "debug", false, "print debug messages")
	f.BoolVar(&globalOpts.quiet, "quiet", false, "print warnings and errors")
}
