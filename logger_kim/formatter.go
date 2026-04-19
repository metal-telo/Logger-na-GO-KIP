package logger_kim

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Formatter formats log entries into strings
type Formatter interface {
	Format(entry Entry) (string, error)
}

// ─── JSON Formatter ──────────────────────────────────────────────────────────

type JSONFormatter struct{ prettyPrint bool }

func NewJSONFormatter() *JSONFormatter       { return &JSONFormatter{} }
func NewPrettyJSONFormatter() *JSONFormatter { return &JSONFormatter{prettyPrint: true} }

func (f *JSONFormatter) Format(entry Entry) (string, error) {
	logEntry := map[string]any{
		"@timestamp": entry.Time.UTC().Format(time.RFC3339Nano),
		"level":      entry.Level.String(),
		"message":    entry.Message,
	}
	for k, v := range entry.Fields {
		logEntry[k] = v
	}
	var data []byte
	var err error
	if f.prettyPrint {
		data, err = json.MarshalIndent(logEntry, "", "  ")
	} else {
		data, err = json.Marshal(logEntry)
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ─── Text Formatter ──────────────────────────────────────────────────────────

type TextFormatter struct {
	timeFormat    string
	includeFields bool
}

func NewTextFormatter() *TextFormatter {
	return &TextFormatter{timeFormat: "2006-01-02 15:04:05", includeFields: true}
}

func (f *TextFormatter) Format(entry Entry) (string, error) {
	var sb strings.Builder
	sb.WriteString(entry.Time.Format(f.timeFormat))
	sb.WriteString(" [")
	sb.WriteString(entry.Level.String())
	sb.WriteString("] ")
	sb.WriteString(entry.Message)
	if f.includeFields && len(entry.Fields) > 0 {
		sb.WriteString(" |")
		for k, v := range entry.Fields {
			sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
		}
	}
	return sb.String(), nil
}

// ─── Pretty Formatter (цветной вывод для разработки) ─────────────────────────

// ColorScheme задаёт 10 цветовых ролей для PrettyFormatter.
// Каждая роль независима и может быть переопределена.
//
// Роли (10 штук):
//  1. Time         — таймстамп
//  2. LevelDebug   — метка DEBUG
//  3. LevelInfo    — метка INFO
//  4. LevelWarn    — метка WARN
//  5. LevelError   — метка ERROR
//  6. LevelFatal   — метка FATAL
//  7. Message      — текст сообщения (обычный)
//  8. MessageError — текст сообщения при Error/Fatal
//  9. FieldKey     — имя поля (ключ)
// 10. FieldValue   — значение поля
type ColorScheme struct {
	Time         *color.Color
	LevelDebug   *color.Color
	LevelInfo    *color.Color
	LevelWarn    *color.Color
	LevelError   *color.Color
	LevelFatal   *color.Color
	Message      *color.Color
	MessageError *color.Color
	FieldKey     *color.Color
	FieldValue   *color.Color
}

// DefaultColorScheme возвращает стандартную схему с 10 цветами.
func DefaultColorScheme() ColorScheme {
	return ColorScheme{
		Time:         color.New(color.FgHiBlack),                     // 1. серый
		LevelDebug:   color.New(color.FgCyan),                        // 2. голубой
		LevelInfo:    color.New(color.FgGreen),                       // 3. зелёный
		LevelWarn:    color.New(color.FgYellow, color.Bold),          // 4. жёлтый жирный
		LevelError:   color.New(color.FgRed),                         // 5. красный
		LevelFatal:   color.New(color.FgHiWhite, color.BgRed, color.Bold), // 6. белый на красном
		Message:      color.New(color.FgHiWhite),                     // 7. яркий белый
		MessageError: color.New(color.FgHiRed, color.Bold),           // 8. яркий красный жирный
		FieldKey:     color.New(color.FgBlue),                        // 9. синий
		FieldValue:   color.New(color.FgHiCyan),                      // 10. яркий голубой
	}
}

// DarkColorScheme — альтернативная схема для тёмных терминалов с высоким контрастом.
func DarkColorScheme() ColorScheme {
	return ColorScheme{
		Time:         color.New(color.FgHiBlack),
		LevelDebug:   color.New(color.FgHiCyan),
		LevelInfo:    color.New(color.FgHiGreen),
		LevelWarn:    color.New(color.FgHiYellow),
		LevelError:   color.New(color.FgHiRed),
		LevelFatal:   color.New(color.FgHiMagenta, color.Bold),
		Message:      color.New(color.FgWhite),
		MessageError: color.New(color.FgHiRed),
		FieldKey:     color.New(color.FgHiBlue),
		FieldValue:   color.New(color.FgHiWhite),
	}
}

type PrettyFormatter struct {
	timeFormat  string
	useColors   bool
	scheme      ColorScheme
}

func NewPrettyFormatter() *PrettyFormatter {
	return &PrettyFormatter{
		timeFormat: "15:04:05.000",
		useColors:  true,
		scheme:     DefaultColorScheme(),
	}
}

// NewPrettyFormatterWithScheme создаёт форматтер с кастомной цветовой схемой.
func NewPrettyFormatterWithScheme(scheme ColorScheme) *PrettyFormatter {
	return &PrettyFormatter{
		timeFormat: "15:04:05.000",
		useColors:  true,
		scheme:     scheme,
	}
}

func (f *PrettyFormatter) levelColor(l Level) *color.Color {
	switch l {
	case LevelDebug:
		return f.scheme.LevelDebug
	case LevelInfo:
		return f.scheme.LevelInfo
	case LevelWarn:
		return f.scheme.LevelWarn
	case LevelError:
		return f.scheme.LevelError
	case LevelFatal:
		return f.scheme.LevelFatal
	default:
		return f.scheme.Message
	}
}

func (f *PrettyFormatter) Format(entry Entry) (string, error) {
	var sb strings.Builder
	s := f.scheme

	// 1. Таймстамп
	ts := entry.Time.Format(f.timeFormat)
	if f.useColors {
		sb.WriteString(s.Time.Sprint(ts))
	} else {
		sb.WriteString(ts)
	}
	sb.WriteString(" ")

	// 2. Уровень (5 ролей)
	levelStr := fmt.Sprintf("%-5s", entry.Level.String())
	if f.useColors {
		sb.WriteString(f.levelColor(entry.Level).Sprint(levelStr))
	} else {
		sb.WriteString(levelStr)
	}
	sb.WriteString(" ")

	// 3. Сообщение (роль 7 или 8 в зависимости от уровня)
	if f.useColors {
		if entry.Level >= LevelError {
			sb.WriteString(s.MessageError.Sprint(entry.Message))
		} else {
			sb.WriteString(s.Message.Sprint(entry.Message))
		}
	} else {
		sb.WriteString(entry.Message)
	}

	// 4. Поля (роли 9 и 10)
	if len(entry.Fields) > 0 {
		sb.WriteString(" ")
		if f.useColors {
			sb.WriteString(s.Time.Sprint("│")) // separator в цвете таймстампа
		} else {
			sb.WriteString("|")
		}
		sb.WriteString(" ")
		first := true
		for k, v := range entry.Fields {
			if !first {
				sb.WriteString(" ")
			}
			first = false
			if f.useColors {
				sb.WriteString(s.FieldKey.Sprint(k))  // ключ
				sb.WriteString("=")
				sb.WriteString(s.FieldValue.Sprint(fmt.Sprint(v))) // значение
			} else {
				sb.WriteString(fmt.Sprintf("%s=%v", k, v))
			}
		}
	}
	return sb.String(), nil
}

// ─── Logfmt Formatter ────────────────────────────────────────────────────────

type LogfmtFormatter struct{ timeFormat string }

func NewLogfmtFormatter() *LogfmtFormatter {
	return &LogfmtFormatter{timeFormat: time.RFC3339}
}

func (f *LogfmtFormatter) Format(entry Entry) (string, error) {
	var sb strings.Builder
	sb.WriteString("time=")
	sb.WriteString(entry.Time.Format(f.timeFormat))
	sb.WriteString(" level=")
	sb.WriteString(entry.Level.String())
	sb.WriteString(" msg=")
	sb.WriteString(quoteIfNeeded(entry.Message))
	for k, v := range entry.Fields {
		sb.WriteString(" ")
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(quoteIfNeeded(fmt.Sprint(v)))
	}
	return sb.String(), nil
}

// ─── Custom Formatter ────────────────────────────────────────────────────────

type CustomFormatter struct {
	formatFunc func(Entry) (string, error)
}

func NewCustomFormatter(fn func(Entry) (string, error)) *CustomFormatter {
	return &CustomFormatter{formatFunc: fn}
}

func (f *CustomFormatter) Format(entry Entry) (string, error) {
	return f.formatFunc(entry)
}

func quoteIfNeeded(s string) string {
	if strings.ContainsAny(s, " \t\n\r\"") {
		return fmt.Sprintf("%q", s)
	}
	return s
}
