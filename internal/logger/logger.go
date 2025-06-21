package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	// ERROR level - only errors
	ERROR LogLevel = iota
	// INFO level - errors and info messages
	INFO
	// DEBUG level - errors, info, and debug messages
	DEBUG
	// VERBOSE level - all messages including verbose output
	VERBOSE
)

// Logger wraps the standard logger with level support
type Logger struct {
	errorLogger   *log.Logger
	infoLogger    *log.Logger
	debugLogger   *log.Logger
	verboseLogger *log.Logger
	level         LogLevel
	logFile       *os.File // Keep reference to close later
}

var defaultLogger *Logger

// Init initializes the global logger with the specified level
func Init(level LogLevel) {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds
	
	defaultLogger = &Logger{
		errorLogger:   log.New(os.Stderr, "ERROR: ", flags),
		infoLogger:    log.New(os.Stdout, "INFO:  ", flags),
		debugLogger:   log.New(os.Stdout, "DEBUG: ", flags),
		verboseLogger: log.New(os.Stdout, "VERBOSE: ", flags),
		level:         level,
	}
}

// InitWithFile initializes the global logger with file and console output
func InitWithFile(level LogLevel, logDir string) error {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds
	
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	
	// Open log file
	logFilePath := filepath.Join(logDir, "score-checker.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	
	// Create multi-writers for both console and file output
	errorWriter := io.MultiWriter(os.Stderr, logFile)
	infoWriter := io.MultiWriter(os.Stdout, logFile)
	
	defaultLogger = &Logger{
		errorLogger:   log.New(errorWriter, "ERROR: ", flags),
		infoLogger:    log.New(infoWriter, "INFO:  ", flags),
		debugLogger:   log.New(infoWriter, "DEBUG: ", flags),
		verboseLogger: log.New(infoWriter, "VERBOSE: ", flags),
		level:         level,
		logFile:       logFile,
	}
	
	return nil
}

// InitFromString initializes the logger from a string level
func InitFromString(levelStr string) {
	level := parseLogLevel(levelStr)
	Init(level)
}

// InitFromStringWithFile initializes the logger from a string level with file output
func InitFromStringWithFile(levelStr string, logDir string) error {
	level := parseLogLevel(levelStr)
	return InitWithFile(level, logDir)
}

// Close closes the log file if it was opened
func Close() error {
	if defaultLogger != nil && defaultLogger.logFile != nil {
		return defaultLogger.logFile.Close()
	}
	return nil
}

// parseLogLevel converts string to LogLevel
func parseLogLevel(levelStr string) LogLevel {
	switch strings.ToUpper(levelStr) {
	case "ERROR":
		return ERROR
	case "INFO":
		return INFO
	case "DEBUG":
		return DEBUG
	case "VERBOSE":
		return VERBOSE
	default:
		return INFO // default to INFO level
	}
}

// SetOutput sets the output destination for all loggers
func SetOutput(w io.Writer) {
	if defaultLogger != nil {
		defaultLogger.errorLogger.SetOutput(w)
		defaultLogger.infoLogger.SetOutput(w)
		defaultLogger.debugLogger.SetOutput(w)
		defaultLogger.verboseLogger.SetOutput(w)
	}
}

// Error logs an error message
func Error(v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.errorLogger.Println(v...)
	}
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.errorLogger.Printf(format, v...)
	}
}

// Info logs an info message
func Info(v ...interface{}) {
	if defaultLogger != nil && defaultLogger.level >= INFO {
		defaultLogger.infoLogger.Println(v...)
	}
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	if defaultLogger != nil && defaultLogger.level >= INFO {
		defaultLogger.infoLogger.Printf(format, v...)
	}
}

// Debug logs a debug message
func Debug(v ...interface{}) {
	if defaultLogger != nil && defaultLogger.level >= DEBUG {
		defaultLogger.debugLogger.Println(v...)
	}
}

// Debugf logs a formatted debug message
func Debugf(format string, v ...interface{}) {
	if defaultLogger != nil && defaultLogger.level >= DEBUG {
		defaultLogger.debugLogger.Printf(format, v...)
	}
}

// Verbose logs a verbose message
func Verbose(v ...interface{}) {
	if defaultLogger != nil && defaultLogger.level >= VERBOSE {
		defaultLogger.verboseLogger.Println(v...)
	}
}

// Verbosef logs a formatted verbose message
func Verbosef(format string, v ...interface{}) {
	if defaultLogger != nil && defaultLogger.level >= VERBOSE {
		defaultLogger.verboseLogger.Printf(format, v...)
	}
}

// GetLevel returns the current log level as a string
func GetLevel() string {
	if defaultLogger == nil {
		return "INFO"
	}
	
	switch defaultLogger.level {
	case ERROR:
		return "ERROR"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case VERBOSE:
		return "VERBOSE"
	default:
		return "INFO"
	}
}