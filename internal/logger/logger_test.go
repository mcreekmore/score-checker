package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogLevel
	}{
		{"ERROR level", "ERROR", ERROR},
		{"INFO level", "INFO", INFO},
		{"DEBUG level", "DEBUG", DEBUG},
		{"VERBOSE level", "VERBOSE", VERBOSE},
		{"lowercase info", "info", INFO},
		{"mixed case debug", "Debug", DEBUG},
		{"invalid level defaults to INFO", "INVALID", INFO},
		{"empty string defaults to INFO", "", INFO},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInitFromString(t *testing.T) {
	// Test initialization
	InitFromString("DEBUG")

	if defaultLogger == nil {
		t.Fatal("defaultLogger should not be nil after initialization")
	}

	if defaultLogger.level != DEBUG {
		t.Errorf("expected log level DEBUG, got %v", defaultLogger.level)
	}
}

func TestLoggerOutput(t *testing.T) {
	// Set up a buffer to capture output
	var buf bytes.Buffer

	// Initialize logger with VERBOSE level to capture all messages
	Init(VERBOSE)
	SetOutput(&buf)

	// Test different log levels
	Error("error message")
	Info("info message")
	Debug("debug message")
	Verbose("verbose message")

	output := buf.String()

	// Check that all messages are present
	if !strings.Contains(output, "error message") {
		t.Error("expected error message in output")
	}
	if !strings.Contains(output, "info message") {
		t.Error("expected info message in output")
	}
	if !strings.Contains(output, "debug message") {
		t.Error("expected debug message in output")
	}
	if !strings.Contains(output, "verbose message") {
		t.Error("expected verbose message in output")
	}

	// Check that log levels are included
	if !strings.Contains(output, "ERROR:") {
		t.Error("expected ERROR: prefix in output")
	}
	if !strings.Contains(output, "INFO:") {
		t.Error("expected INFO: prefix in output")
	}
	if !strings.Contains(output, "DEBUG:") {
		t.Error("expected DEBUG: prefix in output")
	}
	if !strings.Contains(output, "VERBOSE:") {
		t.Error("expected VERBOSE: prefix in output")
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name          string
		level         LogLevel
		expectError   bool
		expectInfo    bool
		expectDebug   bool
		expectVerbose bool
	}{
		{"ERROR level", ERROR, true, false, false, false},
		{"INFO level", INFO, true, true, false, false},
		{"DEBUG level", DEBUG, true, true, true, false},
		{"VERBOSE level", VERBOSE, true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Init(tt.level)
			SetOutput(&buf)

			Error("error")
			Info("info")
			Debug("debug")
			Verbose("verbose")

			output := buf.String()

			hasError := strings.Contains(output, "error")
			hasInfo := strings.Contains(output, "info")
			hasDebug := strings.Contains(output, "debug")
			hasVerbose := strings.Contains(output, "verbose")

			if hasError != tt.expectError {
				t.Errorf("error message presence: got %v, want %v", hasError, tt.expectError)
			}
			if hasInfo != tt.expectInfo {
				t.Errorf("info message presence: got %v, want %v", hasInfo, tt.expectInfo)
			}
			if hasDebug != tt.expectDebug {
				t.Errorf("debug message presence: got %v, want %v", hasDebug, tt.expectDebug)
			}
			if hasVerbose != tt.expectVerbose {
				t.Errorf("verbose message presence: got %v, want %v", hasVerbose, tt.expectVerbose)
			}
		})
	}
}

func TestFormattedLogging(t *testing.T) {
	var buf bytes.Buffer
	Init(VERBOSE)
	SetOutput(&buf)

	Errorf("error: %s", "formatted")
	Infof("info: %d", 42)
	Debugf("debug: %v", true)
	Verbosef("verbose: %s-%d", "test", 123)

	output := buf.String()

	expectedStrings := []string{
		"error: formatted",
		"info: 42",
		"debug: true",
		"verbose: test-123",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected %q in output, got: %s", expected, output)
		}
	}
}

func TestGetLevel(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{ERROR, "ERROR"},
		{INFO, "INFO"},
		{DEBUG, "DEBUG"},
		{VERBOSE, "VERBOSE"},
	}

	for _, tt := range tests {
		Init(tt.level)
		result := GetLevel()
		if result != tt.expected {
			t.Errorf("GetLevel() = %q, want %q", result, tt.expected)
		}
	}
}

func TestUninitializedLogger(t *testing.T) {
	// Reset to nil to test uninitialized state
	defaultLogger = nil

	// These should not panic
	Error("test")
	Info("test")
	Debug("test")
	Verbose("test")

	level := GetLevel()
	if level != "INFO" {
		t.Errorf("GetLevel() for uninitialized logger = %q, want %q", level, "INFO")
	}
}

func TestInitWithFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Initialize logger with file output
	err := InitWithFile(INFO, tempDir)
	if err != nil {
		t.Fatalf("InitWithFile failed: %v", err)
	}

	// Verify logger was created
	if defaultLogger == nil {
		t.Fatal("defaultLogger should not be nil after InitWithFile")
	}

	if defaultLogger.logFile == nil {
		t.Fatal("logFile should not be nil after InitWithFile")
	}

	// Test logging to file
	Info("test file logging")

	// Close the logger
	err = Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Verify log file was created and contains content
	logFilePath := tempDir + "/score-checker.log"
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test file logging") {
		t.Errorf("Log file should contain test message, got: %s", string(content))
	}

	if !strings.Contains(string(content), "INFO:") {
		t.Errorf("Log file should contain INFO prefix, got: %s", string(content))
	}
}

func TestInitFromStringWithFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test with different log levels
	tests := []string{"ERROR", "INFO", "DEBUG", "VERBOSE"}

	for _, level := range tests {
		t.Run(level, func(t *testing.T) {
			err := InitFromStringWithFile(level, tempDir)
			if err != nil {
				t.Fatalf("InitFromStringWithFile failed for level %s: %v", level, err)
			}

			if defaultLogger == nil {
				t.Fatal("defaultLogger should not be nil")
			}

			if GetLevel() != level {
				t.Errorf("expected level %s, got %s", level, GetLevel())
			}

			// Clean up
			Close()
		})
	}
}

func TestFileLoggingPermissionDenied(t *testing.T) {
	// Try to initialize with a non-existent directory that we can't create
	err := InitWithFile(INFO, "/nonexistent/readonly/path")
	if err == nil {
		t.Error("InitWithFile should fail with permission denied")
	}
}
