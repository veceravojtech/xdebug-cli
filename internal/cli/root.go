package cli

import (
	"fmt"
	"os"

	"github.com/console/xdebug-cli/internal/cfg"
	"github.com/spf13/cobra"
)

var (
	// Version is set via ldflags
	Version = "dev"
	// BuildTime is set via ldflags
	BuildTime = "unknown"
)

// CLIArgs holds global command-line parameters
var CLIArgs cfg.CLIParameter

var rootCmd = &cobra.Command{
	Use:   "xdebug-cli",
	Short: "A CLI tool for PHP debugging with Xdebug/DBGp protocol",
	Long:  `xdebug-cli is a command-line DBGp client for debugging PHP applications via Xdebug.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("xdebug-cli version %s (built %s)\n", cfg.Version, BuildTime)
	},
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&CLIArgs.Host, "host", "l", "0.0.0.0", "Host address to listen on for Xdebug connections")
	rootCmd.PersistentFlags().IntVarP(&CLIArgs.Port, "port", "p", 9003, "Port number to listen on for Xdebug connections")
	rootCmd.PersistentFlags().BoolVar(&CLIArgs.JSON, "json", true, "Output results in JSON format")

	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
