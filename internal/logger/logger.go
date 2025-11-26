package logger

import (
	"log"
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
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue  = "\033[34m"
)

func colorForLevel(level Level) string {
	switch level {
	case LevelDebug:
		return colorBlue     // Blue is common for debug
	case LevelInfo:
		return colorGreen    // Green for OK info
	case LevelWarn:
		return colorYellow   // Yellow for warnings
	case LevelError:
		return colorRed      // Red for errors
	default:
		return colorReset
	}
}

func logf(level Level, format string, args ...any) {
	color := colorForLevel(level)
	prefix := color + "[" + string(level) + "] " + colorReset
	log.Printf(prefix+format, args...)
}

func Debugf(format string, args ...any) { logf(LevelDebug, format, args...) }
func Infof(format string, args ...any)  { logf(LevelInfo, format, args...) }
func Warnf(format string, args ...any)  { logf(LevelWarn, format, args...) }
func Errorf(format string, args ...any) { logf(LevelError, format, args...) }
