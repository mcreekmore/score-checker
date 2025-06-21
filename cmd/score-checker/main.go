package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"score-checker/internal/app"
	"score-checker/internal/config"
	"score-checker/internal/logger"
)

var rootCmd = &cobra.Command{
	Use:   "score-checker",
	Short: "Check Sonarr/Radarr episodes/movies for low custom format scores",
	Long:  `A microservice that checks Sonarr episodes and Radarr movies for low custom format scores and optionally triggers searches for better versions.`,
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Close()
		app.RunOnce()
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run as a daemon with periodic checks",
	Long:  `Run the score checker as a daemon that performs periodic checks at the configured interval.`,
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Close()
		
		// Set up signal handling for graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			logger.Info("Received shutdown signal, closing log file...")
			logger.Close()
			os.Exit(0)
		}()
		
		app.RunDaemon()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)

	// Add flags
	rootCmd.PersistentFlags().Bool("triggersearch", false, "Trigger searches for better versions")
	rootCmd.PersistentFlags().Int("batchsize", 5, "Number of items to check per run")
	rootCmd.PersistentFlags().String("interval", "1h", "Interval for daemon mode (e.g., 30m, 1h, 2h30m)")
	rootCmd.PersistentFlags().String("loglevel", "INFO", "Log level (ERROR, INFO, DEBUG, VERBOSE)")

	// Bind flags to viper
	_ = viper.BindPFlag("triggersearch", rootCmd.PersistentFlags().Lookup("triggersearch"))
	_ = viper.BindPFlag("batchsize", rootCmd.PersistentFlags().Lookup("batchsize"))
	_ = viper.BindPFlag("interval", rootCmd.PersistentFlags().Lookup("interval"))
	_ = viper.BindPFlag("loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))
}

func main() {
	config.Init()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
