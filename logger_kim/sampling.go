package logger_kim

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// SamplingHandler — обёртка, пропускающая только часть записей одного уровня.
//
// Стратегия: первые N сообщений в окне времени пропускаются все,
// далее — каждое M-е (thereafter). ERROR и FATAL никогда не семплируются.
//
// Использование:
//
//	base := logger_kim.NewStdoutHandlerWithFormat(logger_kim.FormatPretty)
//	sampled := logger_kim.NewSamplingHandler(base, logger_kim.SamplingConfig{
//	    Tick:       time.Second,
//	    First:      10,   // первые 10/сек — всегда
//	    Thereafter: 100,  // потом каждое 100-е
//	})
type SamplingHandler struct {
	inner  Handler
	config SamplingConfig
	mu     sync.Mutex
	// Счётчики по уровням, сбрасываются каждый Tick
	counters [5]uint64 // индекс = Level + 1 (Debug=-1 → 0, Info=0 → 1 …)
	lastReset time.Time
}

// SamplingConfig описывает политику семплирования
type SamplingConfig struct {
	// Tick — период сброса счётчиков. По умолчанию 1 секунда.
	Tick time.Duration

	// First — количество сообщений каждого уровня за Tick, которые пропускаются без пропуска.
	First uint64

	// Thereafter — каждое N-е сообщение после First тоже пропускается.
	// Остальные отбрасываются.
	Thereafter uint64

	// AlwaysPass — уровни, которые НИКОГДА не семплируются (всегда пишутся).
	// По умолчанию: [LevelError, LevelFatal]
	AlwaysPass []Level
}

// NewSamplingHandler создаёт SamplingHandler
func NewSamplingHandler(inner Handler, cfg SamplingConfig) *SamplingHandler {
	if cfg.Tick <= 0 {
		cfg.Tick = time.Second
	}
	if cfg.First == 0 {
		cfg.First = 10
	}
	if cfg.Thereafter == 0 {
		cfg.Thereafter = 100
	}
	if len(cfg.AlwaysPass) == 0 {
		cfg.AlwaysPass = []Level{LevelError, LevelFatal}
	}
	return &SamplingHandler{
		inner:     inner,
		config:    cfg,
		lastReset: time.Now(),
	}
}

func (h *SamplingHandler) levelIndex(l Level) int {
	// LevelDebug=-1 → 0, Info=0 → 1, Warn=1 → 2, Error=2 → 3, Fatal=3 → 4
	idx := int(l) + 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(h.counters) {
		idx = len(h.counters) - 1
	}
	return idx
}

func (h *SamplingHandler) shouldLog(level Level) bool {
	// Проверяем AlwaysPass
	for _, al := range h.config.AlwaysPass {
		if level >= al {
			return true
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	if now.Sub(h.lastReset) >= h.config.Tick {
		// Сбрасываем все счётчики
		for i := range h.counters {
			h.counters[i] = 0
		}
		h.lastReset = now
	}

	idx := h.levelIndex(level)
	atomic.AddUint64(&h.counters[idx], 1)
	n := atomic.LoadUint64(&h.counters[idx])

	if n <= h.config.First {
		return true
	}
	// После First — пропускаем каждое Thereafter-е
	return (n-h.config.First)%h.config.Thereafter == 0
}

func (h *SamplingHandler) Handle(ctx context.Context, entry Entry) error {
	if !h.shouldLog(entry.Level) {
		return nil
	}
	return h.inner.Handle(ctx, entry)
}

func (h *SamplingHandler) Close() error    { return h.inner.Close() }
func (h *SamplingHandler) Enabled(l Level) bool { return h.inner.Enabled(l) }
func (h *SamplingHandler) SetLevel(l Level)     { h.inner.SetLevel(l) }

// ─── ProbabilisticHandler — вероятностное семплирование ──────────────────────
// Более простая альтернатива: каждое сообщение пропускается с вероятностью Rate.
// Rate=1.0 — пишем всё, Rate=0.1 — пишем ~10%.

type ProbabilisticHandler struct {
	inner Handler
	rate  float64
	rng   *rand.Rand
	mu    sync.Mutex
}

// NewProbabilisticHandler создаёт семплер с заданной вероятностью (0.0–1.0).
// ERROR и FATAL всегда пишутся независимо от rate.
func NewProbabilisticHandler(inner Handler, rate float64) *ProbabilisticHandler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &ProbabilisticHandler{
		inner: inner,
		rate:  rate,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (h *ProbabilisticHandler) Handle(ctx context.Context, entry Entry) error {
	// Критичные уровни — всегда пишем
	if entry.Level >= LevelError {
		return h.inner.Handle(ctx, entry)
	}
	h.mu.Lock()
	pass := h.rng.Float64() < h.rate
	h.mu.Unlock()
	if !pass {
		return nil
	}
	return h.inner.Handle(ctx, entry)
}

func (h *ProbabilisticHandler) Close() error        { return h.inner.Close() }
func (h *ProbabilisticHandler) Enabled(l Level) bool { return h.inner.Enabled(l) }
func (h *ProbabilisticHandler) SetLevel(l Level)     { h.inner.SetLevel(l) }
