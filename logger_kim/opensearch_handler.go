package logger_kim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// OpenSearchHandler отправляет логи в OpenSearch через Bulk API (пакетами)
type OpenSearchHandler struct {
	BaseHandler
	client      *http.Client
	url         string
	index       string
	mu          sync.Mutex
	bufferSize  int
	buffer      []Entry
	flushTicker *time.Ticker
	stopChan    chan struct{}
	async       bool
}

type OpenSearchConfig struct {
	URL           string
	Index         string
	BufferSize    int
	FlushInterval time.Duration
	Async         bool
	Timeout       time.Duration
	Level         Level
}

func NewOpenSearchHandler(url, index string) (*OpenSearchHandler, error) {
	return NewOpenSearchHandlerWithConfig(OpenSearchConfig{
		URL:           url,
		Index:         index,
		BufferSize:    100,
		FlushInterval: 5 * time.Second,
		Async:         true,
		Timeout:       10 * time.Second,
		Level:         LevelInfo,
	})
}

func NewOpenSearchHandlerWithConfig(cfg OpenSearchConfig) (*OpenSearchHandler, error) {
	h := &OpenSearchHandler{
		BaseHandler: NewBaseHandler(cfg.Level, FormatJSON),
		client:      &http.Client{Timeout: cfg.Timeout},
		url:         cfg.URL,
		index:       cfg.Index,
		bufferSize:  cfg.BufferSize,
		buffer:      make([]Entry, 0, cfg.BufferSize),
		stopChan:    make(chan struct{}),
		async:       cfg.Async,
	}

	if err := h.healthCheck(); err != nil {
		return nil, fmt.Errorf("OpenSearch health check failed: %w", err)
	}
	if err := h.createIndex(); err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch index: %w", err)
	}

	if cfg.FlushInterval > 0 {
		h.flushTicker = time.NewTicker(cfg.FlushInterval)
		go h.flushLoop()
	}
	return h, nil
}

func (h *OpenSearchHandler) Handle(ctx context.Context, entry Entry) error {
	if !h.Enabled(entry.Level) {
		return nil
	}
	if h.async {
		go h.sendEntry(ctx, entry)
		return nil
	}
	return h.sendEntry(ctx, entry)
}

func (h *OpenSearchHandler) sendEntry(ctx context.Context, entry Entry) error {
	h.mu.Lock()
	h.buffer = append(h.buffer, entry)
	shouldFlush := len(h.buffer) >= h.bufferSize
	h.mu.Unlock()

	if shouldFlush {
		return h.Flush(ctx)
	}
	return nil
}

func (h *OpenSearchHandler) Flush(ctx context.Context) error {
	h.mu.Lock()
	if len(h.buffer) == 0 {
		h.mu.Unlock()
		return nil
	}
	entries := make([]Entry, len(h.buffer))
	copy(entries, h.buffer)
	h.buffer = h.buffer[:0]
	h.mu.Unlock()
	return h.bulkSend(ctx, entries)
}

func (h *OpenSearchHandler) bulkSend(ctx context.Context, entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	var body bytes.Buffer
	for _, entry := range entries {
		action := map[string]any{"index": map[string]any{"_index": h.index}}
		actionJSON, _ := json.Marshal(action)
		body.Write(actionJSON)
		body.WriteByte('\n')
		doc := h.entryToDocument(entry)
		docJSON, _ := json.Marshal(doc)
		body.Write(docJSON)
		body.WriteByte('\n')
	}
	url := fmt.Sprintf("%s/_bulk", h.url)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-ndjson")
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OpenSearch bulk failed: %d - %s", resp.StatusCode, string(b))
	}
	return nil
}

func (h *OpenSearchHandler) entryToDocument(entry Entry) map[string]any {
	doc := map[string]any{
		"@timestamp": entry.Time.UTC().Format(time.RFC3339Nano),
		"level":      entry.Level.String(),
		"message":    entry.Message,
	}
	for k, v := range entry.Fields {
		doc[k] = v
	}
	return doc
}

func (h *OpenSearchHandler) healthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := fmt.Sprintf("%s/_cluster/health", h.url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("unhealthy status: %d", resp.StatusCode)
	}
	return nil
}

func (h *OpenSearchHandler) createIndex() error {
	url := fmt.Sprintf("%s/%s", h.url, h.index)
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	mapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"@timestamp": map[string]string{"type": "date"},
				"level":      map[string]string{"type": "keyword"},
				"message":    map[string]string{"type": "text"},
			},
		},
	}
	mappingJSON, _ := json.Marshal(mapping)
	req, err = http.NewRequest("PUT", url, bytes.NewReader(mappingJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create index: %d - %s", resp.StatusCode, string(b))
	}
	return nil
}

func (h *OpenSearchHandler) flushLoop() {
	for {
		select {
		case <-h.flushTicker.C:
			h.Flush(context.Background())
		case <-h.stopChan:
			h.Flush(context.Background())
			return
		}
	}
}

func (h *OpenSearchHandler) Close() error {
	if h.flushTicker != nil {
		h.flushTicker.Stop()
	}
	close(h.stopChan)
	return nil
}
