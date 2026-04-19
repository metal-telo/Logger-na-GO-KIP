package logger_kim

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AsyncHandler — обёртка над любым Handler, добавляющая асинхронную буферизацию.
//
// Записи помещаются в канал-буфер и доставляются реальному handler-у
// в отдельной горутине. Это устраняет блокировку вызывающего кода на I/O.
//
// Использование:
//
//	fileH := logger_kim.NewFileHandler("app.log")
//	asyncH := logger_kim.NewAsyncHandler(fileH,
//	    logger_kim.AsyncOptions{BufferSize: 4096, FlushTimeout: 5*time.Second})
//	log := logger_kim.New(logger_kim.Config{Outputs: []logger_kim.Handler{asyncH}})
//	defer log.Close() // обязательно — флашит буфер перед выходом
type AsyncHandler struct {
	inner    Handler
	ch       chan Entry
	wg       sync.WaitGroup
	once     sync.Once
	stopCh   chan struct{}
	opts     AsyncOptions
	overflow uint64 // счётчик пропущенных записей при переполнении буфера
}

// AsyncOptions настраивает поведение AsyncHandler
type AsyncOptions struct {
	// BufferSize — ёмкость канала (количество Entry). По умолчанию 1024.
	BufferSize int

	// FlushTimeout — максимальное время ожидания Close() для отправки оставшихся
	// записей. По умолчанию 10 секунд.
	FlushTimeout time.Duration

	// DropOnFull — если true, при переполнении буфера запись молча отбрасывается
	// вместо блокирующего ожидания. Рекомендуется для high-throughput сценариев.
	DropOnFull bool
}

// NewAsyncHandler создаёт AsyncHandler с заданными параметрами.
// Если opts == AsyncOptions{} используются значения по умолчанию.
func NewAsyncHandler(inner Handler, opts AsyncOptions) *AsyncHandler {
	if opts.BufferSize <= 0 {
		opts.BufferSize = 1024
	}
	if opts.FlushTimeout <= 0 {
		opts.FlushTimeout = 10 * time.Second
	}
	h := &AsyncHandler{
		inner:  inner,
		ch:     make(chan Entry, opts.BufferSize),
		stopCh: make(chan struct{}),
		opts:   opts,
	}
	h.wg.Add(1)
	go h.worker()
	return h
}

// worker — фоновая горутина, сливающая буфер во внутренний handler
func (h *AsyncHandler) worker() {
	defer h.wg.Done()
	for {
		select {
		case entry, ok := <-h.ch:
			if !ok {
				return
			}
			if err := h.inner.Handle(context.Background(), entry); err != nil {
				fmt.Printf("logger_kim/async: inner handler error: %v\n", err)
			}
		case <-h.stopCh:
			// Дочитываем оставшиеся записи из буфера
			for {
				select {
				case entry := <-h.ch:
					_ = h.inner.Handle(context.Background(), entry)
				default:
					return
				}
			}
		}
	}
}

// Handle помещает запись в буфер (неблокирующий при DropOnFull=true)
func (h *AsyncHandler) Handle(_ context.Context, entry Entry) error {
	if h.opts.DropOnFull {
		select {
		case h.ch <- entry:
		default:
			h.overflow++
		}
		return nil
	}
	// Блокирующий вариант
	h.ch <- entry
	return nil
}

// Close сигнализирует worker-у остановиться и ждёт завершения (с таймаутом)
func (h *AsyncHandler) Close() error {
	var err error
	h.once.Do(func() {
		close(h.stopCh)
		done := make(chan struct{})
		go func() { h.wg.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(h.opts.FlushTimeout):
			fmt.Println("logger_kim/async: flush timeout exceeded")
		}
		err = h.inner.Close()
	})
	return err
}

func (h *AsyncHandler) Enabled(level Level) bool  { return h.inner.Enabled(level) }
func (h *AsyncHandler) SetLevel(level Level)       { h.inner.SetLevel(level) }

// DroppedCount возвращает количество записей, отброшенных из-за переполнения буфера
func (h *AsyncHandler) DroppedCount() uint64 { return h.overflow }

// BufferLen возвращает текущее количество записей в буфере
func (h *AsyncHandler) BufferLen() int { return len(h.ch) }
