package config

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"score-checker/internal/logger"
	"score-checker/internal/types"
)

// Init initializes the configuration system
func Init() {
	viper.SetDefault("triggersearch", false)
	viper.SetDefault("batchsize", 5)
	viper.SetDefault("interval", "1h")
	viper.SetDefault("loglevel", "INFO")

	// Read config from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("SCORECHECK")

	// Try to read config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/score-checker/")

	if err := viper.ReadInConfig(); err != nil {
		// Use fmt.Printf here since logger isn't initialized yet
		fmt.Printf("Config file not read: %v\n", err)
	} else {
		// Use fmt.Printf here since logger isn't initialized yet
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}
}

// Load loads configuration using Viper
func Load() types.Config {
	interval, err := time.ParseDuration(viper.GetString("interval"))
	if err != nil {
		log.Fatalf("Invalid interval format: %v", err)
	}

	logLevel := viper.GetString("loglevel")
	
	// Determine log directory based on config file location
	var logDir string
	if configFile := viper.ConfigFileUsed(); configFile != "" {
		// Use the directory where the config file is located
		logDir = filepath.Dir(configFile)
	} else {
		// Default to current directory if no config file
		logDir = "."
	}
	
	// Initialize logger with file and console output
	if err := logger.InitFromStringWithFile(logLevel, logDir); err != nil {
		// Fallback to console-only logging if file logging fails
		fmt.Printf("Warning: Failed to initialize file logging: %v\n", err)
		logger.InitFromString(logLevel)
	}
	
	config := types.Config{
		TriggerSearch: viper.GetBool("triggersearch"),
		BatchSize:     viper.GetInt("batchsize"),
		Interval:      interval,
		LogLevel:      logLevel,
	}

	// Load Sonarr instances
	var sonarrConfig []map[string]interface{}
	if err := viper.UnmarshalKey("sonarr", &sonarrConfig); err == nil {
		for i, instance := range sonarrConfig {
			name, ok := instance["name"].(string)
			if !ok {
				name = "default"
				if i > 0 {
					name = "instance" + string(rune('0'+i))
				}
			}

			baseURL, ok := instance["baseurl"].(string)
			if !ok {
				log.Fatalf("Sonarr instance '%s' missing baseurl", name)
			}

			apiKey, ok := instance["apikey"].(string)
			if !ok {
				log.Fatalf("Sonarr instance '%s' missing apikey", name)
			}

			config.SonarrInstances = append(config.SonarrInstances, types.ServiceConfig{
				Name:    name,
				BaseURL: baseURL,
				APIKey:  apiKey,
			})
		}
	}

	// Load Radarr instances
	var radarrConfig []map[string]interface{}
	if err := viper.UnmarshalKey("radarr", &radarrConfig); err == nil {
		for i, instance := range radarrConfig {
			name, ok := instance["name"].(string)
			if !ok {
				name = "default"
				if i > 0 {
					name = "instance" + string(rune('0'+i))
				}
			}

			baseURL, ok := instance["baseurl"].(string)
			if !ok {
				log.Fatalf("Radarr instance '%s' missing baseurl", name)
			}

			apiKey, ok := instance["apikey"].(string)
			if !ok {
				log.Fatalf("Radarr instance '%s' missing apikey", name)
			}

			config.RadarrInstances = append(config.RadarrInstances, types.ServiceConfig{
				Name:    name,
				BaseURL: baseURL,
				APIKey:  apiKey,
			})
		}
	}

	return config
}
