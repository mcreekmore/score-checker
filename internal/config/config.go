package config

import (
	"log"
	"time"

	"github.com/spf13/viper"

	"score-checker/internal/types"
)

// Init initializes the configuration system
func Init() {
	viper.SetDefault("triggersearch", false)
	viper.SetDefault("batchsize", 5)
	viper.SetDefault("interval", "1h")

	// Read config from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("SCORECHECK")

	// Try to read config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/score-checker/")
	viper.ReadInConfig() // ignore error
}

// Load loads configuration using Viper
func Load() types.Config {
	interval, err := time.ParseDuration(viper.GetString("interval"))
	if err != nil {
		log.Fatalf("Invalid interval format: %v", err)
	}

	config := types.Config{
		TriggerSearch: viper.GetBool("triggersearch"),
		BatchSize:     viper.GetInt("batchsize"),
		Interval:      interval,
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