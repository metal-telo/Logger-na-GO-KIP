package logger_kim

// Config holds the configuration for the logger
type Config struct {
	// Level is the minimum log level to output
	Level Level

	// Format is the output format (JSON, Text, Pretty)
	Format Format

	// Outputs is a list of handlers to send logs to
	Outputs []Handler

	// AddSource adds source code location (file, line) to logs
	AddSource bool

	// TimeFormat specifies the time format for logs
	TimeFormat string

	// ServiceName is added to all log entries as "service" field
	ServiceName string

	// Environment is added to all log entries as "environment" field
	Environment string

	// AdditionalFields are added to every log entry
	AdditionalFields map[string]any

	// Features декларативно включает или отключает отдельные возможности
	// библиотеки: Async, Sampling, SlogBridge, PrettyColor, FileOutput и др.
	// Нулевое значение (Features{}) — минимальный синхронный логгер.
	// Используйте DefaultFeatures(), ProductionFeatures() или свой набор.
	Features Features
}

// Level represents the severity level of a log
type Level int

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel converts a string to a Level
func ParseLevel(s string) Level {
	switch s {
	case "DEBUG", "debug":
		return LevelDebug
	case "INFO", "info":
		return LevelInfo
	case "WARN", "warn", "WARNING", "warning":
		return LevelWarn
	case "ERROR", "error":
		return LevelError
	case "FATAL", "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// Format represents the output format
type Format string

const (
	FormatJSON   Format = "json"
	FormatText   Format = "text"
	FormatPretty Format = "pretty"
	FormatLogfmt Format = "logfmt"
)

// DefaultConfig returns a sensible default configuration
func DefaultConfig() Config {
	return Config{
		Level:  LevelInfo,
		Format: FormatJSON,
	}
}
