package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"score-checker/internal/app"
	"score-checker/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "score-checker",
	Short: "Check Sonarr/Radarr episodes/movies for low custom format scores",
	Long:  `A microservice that checks Sonarr episodes and Radarr movies for low custom format scores and optionally triggers searches for better versions.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.RunOnce()
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run as a daemon with periodic checks",
	Long:  `Run the score checker as a daemon that performs periodic checks at the configured interval.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.RunDaemon()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)

	// Add flags
	rootCmd.PersistentFlags().Bool("triggersearch", false, "Trigger searches for better versions")
	rootCmd.PersistentFlags().Int("batchsize", 5, "Number of items to check per run")
	rootCmd.PersistentFlags().String("interval", "1h", "Interval for daemon mode (e.g., 30m, 1h, 2h30m)")

	// Bind flags to viper
	_ = viper.BindPFlag("triggersearch", rootCmd.PersistentFlags().Lookup("triggersearch"))
	_ = viper.BindPFlag("batchsize", rootCmd.PersistentFlags().Lookup("batchsize"))
	_ = viper.BindPFlag("interval", rootCmd.PersistentFlags().Lookup("interval"))
}

func main() {
	config.Init()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
