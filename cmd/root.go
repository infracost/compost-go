package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "compost",
	Short: "Post pull request comments from multiple CI platforms",
	Long:  "Post pull request comments from multiple CI platforms",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ctx, err := loggerContext(rootCmd)
	if err != nil {
		handleErr(ctx, err)
	}

	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		handleErr(ctx, err)
	}
}

// handleErr logs the error and exits the program.
func handleErr(ctx context.Context, err error) {
	log.Ctx(ctx).Error().Msgf(err.Error())
	os.Exit(1)
}

// defaultLogger returns a new logger with the default settings.
func defaultLogger(cmd *cobra.Command) (zerolog.Logger, error) {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	logLevelVal, _ := cmd.Flags().GetString("log-level")
	if logLevelVal == "" {
		logLevelVal = "info"
	}

	logLevel, err := zerolog.ParseLevel(logLevelVal)
	if err != nil {
		logLevel = zerolog.InfoLevel
		err = fmt.Errorf("Invalid log level '%s'", logLevelVal)
	}

	return zerolog.New(output).Level(logLevel).With().Timestamp().Logger(), err
}

// loggerContext returns a new context with the default logger attached to it.
func loggerContext(cmd *cobra.Command) (context.Context, error) {
	ctx := context.Background()
	log, err := defaultLogger(cmd)
	if err != nil {
		return nil, err
	}
	ctx = log.WithContext(ctx)
	return ctx, nil
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().String("log-level", "", "Log level: trace, debug, info, warn, error, fatal")
}
