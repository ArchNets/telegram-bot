package logger

import (
	"log"
	"strings"

	"github.com/archnets/telegram-bot/internal/env"
)

type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

// levelPriority returns the numeric priority for a log level.
// Higher number = more severe.
func levelPriority(level Level) int {
	switch level {
	case LevelDebug:
		return 0
	case LevelInfo:
		return 1
	case LevelWarn:
		return 2
	case LevelError:
		return 3
	default:
		return 1 // Default to INFO
	}
}

// getConfiguredLevel returns the configured log level from environment.
func getConfiguredLevel() Level {
	val := env.GetString("LOG_LEVEL", "INFO")
	switch strings.ToUpper(val) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

func colorForLevel(level Level) string {
	switch level {
	case LevelDebug:
		return colorBlue // Blue for debug
	case LevelInfo:
		return colorGreen // Green for info
	case LevelWarn:
		return colorYellow // Yellow for warnings
	case LevelError:
		return colorRed // Red for errors
	default:
		return colorReset
	}
}

func logf(level Level, format string, args ...any) {
	// Skip if below configured level
	if levelPriority(level) < levelPriority(getConfiguredLevel()) {
		return
	}

	color := colorForLevel(level)
	prefix := color + "[" + string(level) + "] " + colorReset
	log.Printf(prefix+format, args...)
}

func Debugf(format string, args ...any) { logf(LevelDebug, format, args...) }
func Infof(format string, args ...any)  { logf(LevelInfo, format, args...) }
func Warnf(format string, args ...any)  { logf(LevelWarn, format, args...) }
func Errorf(format string, args ...any) { logf(LevelError, format, args...) }
