package logger_kim

import "time"

// Features — декларативный набор флагов, включающих или отключающих отдельные
// возможности logger_kim. Передаётся в Config.Features.
//
// Каждый флаг независим. По умолчанию всё отключено (Features{}) —
// минимальный синхронный логгер без сторонних оберток.
//
// Порядок оборачивания handlers (снаружи → внутрь):
//
//	SamplingHandler → AsyncHandler → оригинальный Handler
//
// Это значит: Sampling фильтрует первым, AsyncHandler доставляет в фоне,
// реальный I/O — последний.
type Features struct {
	// Async оборачивает каждый Handler из Config.Outputs в AsyncHandler.
	// Вызываемый код перестаёт блокироваться на I/O.
	// Default: false.
	Async bool

	// AsyncOpts параметры для AsyncHandler (применимо когда Async == true).
	// Нулевые значения заменяются дефолтами (BufferSize=1024, FlushTimeout=10s).
	AsyncOpts AsyncOptions

	// Sampling оборачивает каждый Handler в SamplingHandler.
	// Ограничивает поток логов — полезно при высокой нагрузке.
	// Default: false.
	Sampling bool

	// SamplingOpts параметры для SamplingHandler (когда Sampling == true).
	// Нулевые значения заменяются дефолтами (Tick=1s, First=10, Thereafter=100).
	SamplingOpts SamplingConfig

	// SlogBridge автоматически вызывает SetAsDefault(logger) внутри New().
	// После этого slog.Info / log.Println и весь код в сторонних библиотеках
	// идут через этот логгер.
	// Default: false — вызывайте SetAsDefault вручную, если нужна явность.
	SlogBridge bool

	// PrettyColor включает ANSI-цвета в Pretty-форматтере.
	// Отключите, если терминал не поддерживает ANSI или нужен plain-text.
	// Default: true.
	PrettyColor bool

	// ColorScheme задаёт 10 цветовых ролей для PrettyFormatter.
	// Нулевое значение (ColorScheme{}) означает DefaultColorScheme().
	// Используйте DarkColorScheme() или свою схему.
	ColorScheme ColorScheme

	// StructuredFields разрешает структурированные поля key=value / JSON.
	// Отключите для ультра-минимального вывода только с msg.
	// Default: true.
	StructuredFields bool

	// MultipleOutputs разрешает более одного Handler в Config.Outputs.
	// Если false и передано несколько handlers — используется только первый.
	// Позволяет динамически ограничить число targets без изменения Config.
	// Default: true.
	MultipleOutputs bool

	// FileOutput разрешает использование FileHandler.
	// Если false — FileHandler-ы из Outputs молча пропускаются.
	// Default: true.
	FileOutput bool

	// OpenSearchOutput разрешает использование OpenSearchHandler.
	// Default: true.
	OpenSearchOutput bool
}

// DefaultFeatures возвращает Features с включёнными "базовыми" опциями:
// цвет в Pretty, структурированные поля, возможность нескольких outputs.
// Async, Sampling и SlogBridge — выключены (явное is better than implicit).
func DefaultFeatures() Features {
	return Features{
		PrettyColor:      true,
		StructuredFields: true,
		MultipleOutputs:  true,
		FileOutput:       true,
		OpenSearchOutput: true,
	}
}

// MinimalFeatures — нулевой набор зависимостей и возможностей.
// Только stdout, только текст, никаких оберток, никакого slog-bridge.
// Подходит для встроенных CLI-инструментов или тестов.
func MinimalFeatures() Features {
	return Features{
		StructuredFields: true,
		MultipleOutputs:  false,
		FileOutput:       false,
		OpenSearchOutput: false,
	}
}

// ProductionFeatures — рекомендованный набор для production:
// Async + Sampling + file/OpenSearch включены, цвет отключён (JSON-only).
func ProductionFeatures() Features {
	return Features{
		Async: true,
		AsyncOpts: AsyncOptions{
			BufferSize:   4096,
			FlushTimeout: 10 * time.Second,
			DropOnFull:   true,
		},
		Sampling: true,
		SamplingOpts: SamplingConfig{
			Tick:       0, // default 1s
			First:      10,
			Thereafter: 100,
		},
		SlogBridge:       true,
		PrettyColor:      false, // production — JSON, не Pretty
		StructuredFields: true,
		MultipleOutputs:  true,
		FileOutput:       true,
		OpenSearchOutput: true,
	}
}

// ─── internal ─────────────────────────────────────────────────────────────────

// applyFeatures применяет флаги Features к списку handlers:
//   - фильтрует FileHandler / OpenSearchHandler если соответствующий флаг false
//   - ограничивает до первого handler если MultipleOutputs == false
//   - оборачивает в AsyncHandler если Async == true
//   - оборачивает в SamplingHandler если Sampling == true
func applyFeatures(handlers []Handler, f Features) []Handler {
	// 1. Фильтрация по типу handler
	filtered := handlers
	if !f.FileOutput || !f.OpenSearchOutput {
		kept := make([]Handler, 0, len(handlers))
		for _, h := range handlers {
			switch h.(type) {
			case *FileHandler:
				if f.FileOutput {
					kept = append(kept, h)
				}
			case *OpenSearchHandler:
				if f.OpenSearchOutput {
					kept = append(kept, h)
				}
			default:
				kept = append(kept, h)
			}
		}
		filtered = kept
	}

	// 2. Ограничение числа outputs
	if !f.MultipleOutputs && len(filtered) > 1 {
		filtered = filtered[:1]
	}

	// 3. Применить ColorScheme к StdoutHandler с FormatPretty (если PrettyColor=true)
	if f.PrettyColor {
		scheme := f.ColorScheme
		if scheme.LevelInfo == nil { // нулевая схема → используем дефолт
			scheme = DefaultColorScheme()
		}
		for _, h := range filtered {
			if sh, ok := h.(*StdoutHandler); ok && sh.GetFormat() == FormatPretty {
				sh.SetFormatter(NewPrettyFormatterWithScheme(scheme))
			}
		}
	}

	// 4. Нет оберток — возвращаем как есть
	if !f.Async && !f.Sampling {
		return filtered
	}

	// 5. Оборачивание: порядок Sampling(outermost) → Async → Handler
	wrapped := make([]Handler, len(filtered))
	for i, h := range filtered {
		wh := h
		if f.Async {
			opts := f.AsyncOpts
			if opts.BufferSize <= 0 {
				opts.BufferSize = 1024
			}
			wh = NewAsyncHandler(wh, opts)
		}
		if f.Sampling {
			wh = NewSamplingHandler(wh, f.SamplingOpts)
		}
		wrapped[i] = wh
	}
	return wrapped
}
