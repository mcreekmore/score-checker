package config

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	Init()

	// Test default values
	if viper.GetBool("triggersearch") != false {
		t.Error("expected triggersearch default to be false")
	}
	if viper.GetInt("batchsize") != 5 {
		t.Error("expected batchsize default to be 5")
	}
	if viper.GetString("interval") != "1h" {
		t.Error("expected interval default to be '1h'")
	}
}

func TestLoadWithDefaults(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	Init()

	cfg := Load()

	// Test defaults
	if cfg.TriggerSearch != false {
		t.Error("expected TriggerSearch to be false by default")
	}
	if cfg.BatchSize != 5 {
		t.Error("expected BatchSize to be 5 by default")
	}
	if cfg.Interval != time.Hour {
		t.Errorf("expected Interval to be 1 hour, got %v", cfg.Interval)
	}
	if len(cfg.SonarrInstances) != 0 {
		t.Error("expected no Sonarr instances by default")
	}
	if len(cfg.RadarrInstances) != 0 {
		t.Error("expected no Radarr instances by default")
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	// Set environment variables
	os.Setenv("SCORECHECK_TRIGGERSEARCH", "true")
	os.Setenv("SCORECHECK_BATCHSIZE", "10")
	os.Setenv("SCORECHECK_INTERVAL", "30m")

	defer func() {
		os.Unsetenv("SCORECHECK_TRIGGERSEARCH")
		os.Unsetenv("SCORECHECK_BATCHSIZE")
		os.Unsetenv("SCORECHECK_INTERVAL")
	}()

	Init()
	cfg := Load()

	if !cfg.TriggerSearch {
		t.Error("expected TriggerSearch to be true from environment")
	}
	if cfg.BatchSize != 10 {
		t.Errorf("expected BatchSize to be 10 from environment, got %d", cfg.BatchSize)
	}
	if cfg.Interval != 30*time.Minute {
		t.Errorf("expected Interval to be 30 minutes from environment, got %v", cfg.Interval)
	}
}

func TestLoadWithViperConfig(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	Init()

	// Set viper values directly to simulate config file
	viper.Set("triggersearch", true)
	viper.Set("batchsize", 15)
	viper.Set("interval", "2h")

	// Set up Sonarr instances
	sonarrInstances := []map[string]interface{}{
		{
			"name":    "main",
			"baseurl": "http://localhost:8989",
			"apikey":  "test-sonarr-key",
		},
		{
			"name":    "4k",
			"baseurl": "http://localhost:8990",
			"apikey":  "test-sonarr-4k-key",
		},
	}
	viper.Set("sonarr", sonarrInstances)

	// Set up Radarr instances
	radarrInstances := []map[string]interface{}{
		{
			"name":    "main",
			"baseurl": "http://localhost:7878",
			"apikey":  "test-radarr-key",
		},
	}
	viper.Set("radarr", radarrInstances)

	cfg := Load()

	// Test general config
	if !cfg.TriggerSearch {
		t.Error("expected TriggerSearch to be true")
	}
	if cfg.BatchSize != 15 {
		t.Errorf("expected BatchSize to be 15, got %d", cfg.BatchSize)
	}
	if cfg.Interval != 2*time.Hour {
		t.Errorf("expected Interval to be 2 hours, got %v", cfg.Interval)
	}

	// Test Sonarr instances
	if len(cfg.SonarrInstances) != 2 {
		t.Errorf("expected 2 Sonarr instances, got %d", len(cfg.SonarrInstances))
	}
	if cfg.SonarrInstances[0].Name != "main" {
		t.Errorf("expected first Sonarr instance name to be 'main', got %q", cfg.SonarrInstances[0].Name)
	}
	if cfg.SonarrInstances[0].BaseURL != "http://localhost:8989" {
		t.Errorf("expected first Sonarr instance baseurl to be 'http://localhost:8989', got %q", cfg.SonarrInstances[0].BaseURL)
	}
	if cfg.SonarrInstances[0].APIKey != "test-sonarr-key" {
		t.Errorf("expected first Sonarr instance apikey to be 'test-sonarr-key', got %q", cfg.SonarrInstances[0].APIKey)
	}
	if cfg.SonarrInstances[1].Name != "4k" {
		t.Errorf("expected second Sonarr instance name to be '4k', got %q", cfg.SonarrInstances[1].Name)
	}

	// Test Radarr instances
	if len(cfg.RadarrInstances) != 1 {
		t.Errorf("expected 1 Radarr instance, got %d", len(cfg.RadarrInstances))
	}
	if cfg.RadarrInstances[0].Name != "main" {
		t.Errorf("expected Radarr instance name to be 'main', got %q", cfg.RadarrInstances[0].Name)
	}
	if cfg.RadarrInstances[0].BaseURL != "http://localhost:7878" {
		t.Errorf("expected Radarr instance baseurl to be 'http://localhost:7878', got %q", cfg.RadarrInstances[0].BaseURL)
	}
	if cfg.RadarrInstances[0].APIKey != "test-radarr-key" {
		t.Errorf("expected Radarr instance apikey to be 'test-radarr-key', got %q", cfg.RadarrInstances[0].APIKey)
	}
}

func TestLoadWithDefaultInstanceNames(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	Init()

	// Set up instances without names
	sonarrInstances := []map[string]interface{}{
		{
			"baseurl": "http://localhost:8989",
			"apikey":  "test-sonarr-key",
		},
		{
			"baseurl": "http://localhost:8990",
			"apikey":  "test-sonarr-4k-key",
		},
	}
	viper.Set("sonarr", sonarrInstances)

	cfg := Load()

	if len(cfg.SonarrInstances) != 2 {
		t.Errorf("expected 2 Sonarr instances, got %d", len(cfg.SonarrInstances))
	}
	if cfg.SonarrInstances[0].Name != "default" {
		t.Errorf("expected first instance name to be 'default', got %q", cfg.SonarrInstances[0].Name)
	}
	if cfg.SonarrInstances[1].Name != "instance1" {
		t.Errorf("expected second instance name to be 'instance1', got %q", cfg.SonarrInstances[1].Name)
	}
}

func TestLoadInvalidInterval(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	Init()

	viper.Set("interval", "invalid")

	// This should cause a fatal error, but we can't easily test that
	// In a real scenario, this would call log.Fatalf
	// For testing purposes, we'd need to refactor the code to return errors
	// instead of calling log.Fatalf directly
}

func TestLoadMissingRequiredFields(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	Init()

	t.Run("missing sonarr baseurl", func(t *testing.T) {
		sonarrInstances := []map[string]interface{}{
			{
				"name":   "main",
				"apikey": "test-key",
				// missing baseurl
			},
		}
		viper.Set("sonarr", sonarrInstances)

		// This should cause a fatal error in Load()
		// In a real test, we'd need to refactor to return errors
	})

	t.Run("missing sonarr apikey", func(t *testing.T) {
		viper.Reset()
		Init()

		sonarrInstances := []map[string]interface{}{
			{
				"name":    "main",
				"baseurl": "http://localhost:8989",
				// missing apikey
			},
		}
		viper.Set("sonarr", sonarrInstances)

		// This should cause a fatal error in Load()
		// In a real test, we'd need to refactor to return errors
	})
}

func TestLoadEmptyInstanceArrays(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	Init()

	// Set empty arrays
	viper.Set("sonarr", []map[string]interface{}{})
	viper.Set("radarr", []map[string]interface{}{})

	cfg := Load()

	if len(cfg.SonarrInstances) != 0 {
		t.Errorf("expected 0 Sonarr instances, got %d", len(cfg.SonarrInstances))
	}
	if len(cfg.RadarrInstances) != 0 {
		t.Errorf("expected 0 Radarr instances, got %d", len(cfg.RadarrInstances))
	}
}

func TestCustomHandler(t *testing.T) {
	var buf bytes.Buffer
	handler := &customHandler{
		writer: &buf,
		level:  slog.LevelDebug,
	}

	ctx := context.Background()

	// Test Enabled method
	if !handler.Enabled(ctx, slog.LevelInfo) {
		t.Error("expected handler to be enabled for INFO level")
	}
	if !handler.Enabled(ctx, slog.LevelError) {
		t.Error("expected handler to be enabled for ERROR level")
	}
	if !handler.Enabled(ctx, slog.LevelDebug) {
		t.Error("expected handler to be enabled for DEBUG level")
	}

	// Test Handle method
	record := slog.NewRecord(time.Date(2024, 1, 3, 10, 24, 22, 0, time.UTC), slog.LevelInfo, "Test message", 0)
	err := handler.Handle(ctx, record)
	if err != nil {
		t.Errorf("unexpected error handling log record: %v", err)
	}

	output := buf.String()
	expectedParts := []string{
		"2024/01/03 10:24:22",
		"INFO",
		"Test message",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("expected output to contain %q, got: %s", part, output)
		}
	}

	// Test WithAttrs and WithGroup (they just return the handler)
	newHandler := handler.WithAttrs([]slog.Attr{})
	if newHandler != handler {
		t.Error("expected WithAttrs to return the same handler")
	}

	newHandler = handler.WithGroup("test")
	if newHandler != handler {
		t.Error("expected WithGroup to return the same handler")
	}
}

func TestInitLogger(t *testing.T) {
	// This test verifies that initLogger doesn't panic and sets up slog
	originalLogger := slog.Default()
	defer slog.SetDefault(originalLogger)

	initLogger()

	// Test that we can log something without panicking
	slog.Info("test message")
}

func TestInitLoggerWithFile(t *testing.T) {
	// Test successful file logger initialization
	tempDir := t.TempDir()

	err := initLoggerWithFile(tempDir)
	if err != nil {
		t.Errorf("unexpected error initializing logger with file: %v", err)
	}

	// Verify log file was created
	logFile := tempDir + "/score-checker.log"
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("expected log file to be created")
	}

	// Test that we can log to the file
	slog.Info("test log message")

	// Check if the log was written
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Errorf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test log message") {
		t.Error("expected log file to contain test message")
	}
}

func TestInitLoggerWithFileError(t *testing.T) {
	// Test error case - try to create log in non-existent directory without permissions
	err := initLoggerWithFile("/nonexistent/readonly/path")
	if err == nil {
		t.Error("expected error when trying to create log directory with no permissions")
	}
}

// Helper function to reset viper state
func resetViper() {
	viper.Reset()
}

func TestMain(m *testing.M) {
	// Setup
	code := m.Run()
	// Teardown
	resetViper()
	os.Exit(code)
}
