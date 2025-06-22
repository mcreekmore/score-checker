package config

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"score-checker/internal/types"
)

// customHandler implements a simple log format: "2024/01/03 10:24:22 INFO Info message"
type customHandler struct {
	writer io.Writer
	level  slog.Level
}

func (h *customHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *customHandler) Handle(_ context.Context, r slog.Record) error {
	timestamp := r.Time.Format("2006/01/02 15:04:05")
	level := r.Level.String()
	message := r.Message

	line := fmt.Sprintf("%s %s %s\n", timestamp, level, message)
	_, err := h.writer.Write([]byte(line))
	return err
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *customHandler) WithGroup(name string) slog.Handler {
	return h
}

// initLogger initializes slog with console output and custom format
func initLogger() {
	handler := &customHandler{
		writer: os.Stdout,
		level:  slog.LevelDebug,
	}
	slog.SetDefault(slog.New(handler))
}

// initLoggerWithFile initializes slog with file and console output
func initLoggerWithFile(logDir string) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFilePath := filepath.Join(logDir, "score-checker.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	handler := &customHandler{
		writer: multiWriter,
		level:  slog.LevelDebug,
	}
	slog.SetDefault(slog.New(handler))

	return nil
}

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
	if err := initLoggerWithFile(logDir); err != nil {
		// Fallback to console-only logging if file logging fails
		fmt.Printf("Warning: Failed to initialize file logging: %v\n", err)
		initLogger()
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
