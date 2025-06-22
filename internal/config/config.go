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

func parseInterval() time.Duration {
	interval, err := time.ParseDuration(viper.GetString("interval"))
	if err != nil {
		log.Fatalf("Invalid interval format: %v", err)
	}
	return interval
}

func determineLogDir() string {
	if configFile := viper.ConfigFileUsed(); configFile != "" {
		return filepath.Dir(configFile)
	}
	return "."
}

func setupLogging() {
	logDir := determineLogDir()
	if err := initLoggerWithFile(logDir); err != nil {
		fmt.Printf("Warning: Failed to initialize file logging: %v\n", err)
		initLogger()
	}
}

func generateInstanceName(index int) string {
	if index == 0 {
		return "default"
	}
	return "instance" + string(rune('0'+index))
}

func parseServiceInstance(instance map[string]interface{}, index int, serviceName string) types.ServiceConfig {
	name, ok := instance["name"].(string)
	if !ok {
		name = generateInstanceName(index)
	}

	baseURL, ok := instance["baseurl"].(string)
	if !ok {
		log.Fatalf("%s instance '%s' missing baseurl", serviceName, name)
	}

	apiKey, ok := instance["apikey"].(string)
	if !ok {
		log.Fatalf("%s instance '%s' missing apikey", serviceName, name)
	}

	return types.ServiceConfig{
		Name:    name,
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
}

func loadServiceInstances(key, serviceName string) []types.ServiceConfig {
	var instances []types.ServiceConfig
	var serviceConfig []map[string]interface{}

	if err := viper.UnmarshalKey(key, &serviceConfig); err == nil {
		for i, instance := range serviceConfig {
			config := parseServiceInstance(instance, i, serviceName)
			instances = append(instances, config)
		}
	}

	return instances
}

// Load loads configuration using Viper
func Load() types.Config {
	interval := parseInterval()
	setupLogging()

	config := types.Config{
		TriggerSearch: viper.GetBool("triggersearch"),
		BatchSize:     viper.GetInt("batchsize"),
		Interval:      interval,
		LogLevel:      viper.GetString("loglevel"),
	}

	config.SonarrInstances = loadServiceInstances("sonarr", "Sonarr")
	config.RadarrInstances = loadServiceInstances("radarr", "Radarr")

	return config
}
