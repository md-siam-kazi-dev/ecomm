package security

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// writeBatchSize is the row count at which a batch is flushed immediately.
	writeBatchSize = 100
	// writeFlushInterval bounds how long an event waits before being written.
	writeFlushInterval = 250 * time.Millisecond
	// maxUserAgent mirrors the audit trail's 512-byte cap.
	maxUserAgent = 512
)

const insertColumns = `(event_id, request_id, event_type, severity, user_id, user_role, ` +
	`ip_address, method, endpoint, status_code, duration_ms, user_agent, ` +
	`resource_type, resource_id, metadata)`

const insertColumnCount = 15

// Writer persists security events off the request path.
//
// Events are queued into a bounded channel and written in batches by a single
// goroutine, so a slow or unavailable database cannot add latency to an API
// request. When the queue is full the event is dropped and counted — monitoring
// degradation is reported through Dropped(), never hidden.
//
// Critical events can bypass the queue with InsertNow so they are never lost.
type Writer struct {
	db    *pgxpool.Pool
	log   *slog.Logger
	queue chan SecurityEvent

	started sync.Once
	stop    chan struct{}
	done    chan struct{}

	closing atomic.Bool
	dropped atomic.Int64
	written atomic.Int64
}

// NewWriter builds a writer with a bounded queue.
func NewWriter(db *pgxpool.Pool, log *slog.Logger, queueSize int) *Writer {
	if queueSize <= 0 {
		queueSize = 1024
	}
	if log == nil {
		log = slog.Default()
	}
	return &Writer{
		db:    db,
		log:   log,
		queue: make(chan SecurityEvent, queueSize),
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
}

// Start launches the batch writer. It is idempotent.
func (w *Writer) Start() {
	w.started.Do(func() { go w.run() })
}

func (w *Writer) run() {
	defer close(w.done)

	ticker := time.NewTicker(writeFlushInterval)
	defer ticker.Stop()

	batch := make([]SecurityEvent, 0, writeBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		w.flush(ctx, batch)
		cancel()
		batch = batch[:0]
	}

	enqueue := func(ev SecurityEvent) {
		batch = append(batch, ev)
		if len(batch) >= writeBatchSize {
			flush()
		}
	}

	for {
		select {
		case ev := <-w.queue:
			enqueue(ev)
		case <-ticker.C:
			flush()
		case <-w.stop:
			// Drain whatever is already queued, then flush once.
			for {
				select {
				case ev := <-w.queue:
					enqueue(ev)
					continue
				default:
				}
				break
			}
			flush()
			return
		}
	}
}

// Enqueue hands an event to the writer. It never blocks; the bool reports
// whether the event was accepted.
func (w *Writer) Enqueue(ev SecurityEvent) bool {
	if w == nil || w.db == nil || w.closing.Load() {
		return false
	}
	select {
	case w.queue <- ev:
		return true
	default:
		w.dropped.Add(1)
		return false
	}
}

// InsertNow writes a single event synchronously. Used for critical events, whose
// volume is low enough that the extra latency is irrelevant and whose loss would
// not be.
func (w *Writer) InsertNow(ctx context.Context, ev SecurityEvent) error {
	if w == nil || w.db == nil {
		return nil
	}
	sql, args := buildInsert([]SecurityEvent{ev})
	if _, err := w.db.Exec(ctx, sql, args...); err != nil {
		w.log.ErrorContext(ctx, "security event write failed",
			slog.String("event_type", string(ev.Type)),
			slog.String("error", err.Error()),
		)
		return err
	}
	w.written.Add(1)
	return nil
}

func (w *Writer) flush(ctx context.Context, batch []SecurityEvent) {
	if w.db == nil || len(batch) == 0 {
		return
	}
	sql, args := buildInsert(batch)
	if _, err := w.db.Exec(ctx, sql, args...); err != nil {
		w.log.ErrorContext(ctx, "security event batch write failed",
			slog.Int("count", len(batch)),
			slog.String("error", err.Error()),
		)
		return
	}
	w.written.Add(int64(len(batch)))
}

// Close stops accepting events, drains the queue and waits for the final flush.
func (w *Writer) Close(ctx context.Context) {
	if w == nil {
		return
	}
	if !w.closing.CompareAndSwap(false, true) {
		return
	}
	close(w.stop)
	select {
	case <-w.done:
	case <-ctx.Done():
		w.log.WarnContext(ctx, "security event writer did not drain before shutdown")
	}
}

// Dropped is the number of events discarded because the queue was full.
func (w *Writer) Dropped() int64 {
	if w == nil {
		return 0
	}
	return w.dropped.Load()
}

// Written is the number of events persisted successfully (used by tests).
func (w *Writer) Written() int64 {
	if w == nil {
		return 0
	}
	return w.written.Load()
}

func buildInsert(batch []SecurityEvent) (string, []any) {
	var sb strings.Builder
	sb.WriteString("INSERT INTO security_events ")
	sb.WriteString(insertColumns)
	sb.WriteString(" VALUES ")

	args := make([]any, 0, len(batch)*insertColumnCount)
	for i, ev := range batch {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("(")
		for col := 1; col <= insertColumnCount; col++ {
			if col > 1 {
				sb.WriteString(",")
			}
			sb.WriteString("$")
			sb.WriteString(strconv.Itoa(i*insertColumnCount + col))
		}
		sb.WriteString(")")

		args = append(args,
			ev.EventID,
			nullableString(ev.RequestID),
			string(ev.Type),
			string(ev.Severity),
			sqlUUID(ev.UserID),
			nullableString(ev.UserRole),
			sqlIP(ev.IP),
			nullableString(ev.Method),
			nullableString(ev.Endpoint),
			sqlInt(ev.StatusCode),
			ev.DurationMS,
			truncateString(ev.UserAgent, maxUserAgent),
			nullableString(ev.ResourceType),
			sqlUUID(ev.ResourceID),
			sqlJSON(ev.Metadata),
		)
	}

	return sb.String(), args
}

// sqlUUID mirrors the audit recorder: a nil interface for empty/invalid values so
// the UUID column stores NULL instead of failing the whole batch.
func sqlUUID(value string) any {
	if value == "" {
		return nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil
	}
	return id
}

// sqlIP validates the value for the INET column, dropping host:port and garbage.
func sqlIP(value string) any {
	if value == "" {
		return nil
	}
	host := value
	if h, _, err := net.SplitHostPort(value); err == nil {
		host = h
	}
	if net.ParseIP(strings.Trim(host, "[]")) == nil {
		return nil
	}
	return strings.Trim(host, "[]")
}

func sqlInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

// sqlJSON sanitizes and encodes metadata. A value that cannot be marshalled is
// dropped rather than failing the batch.
func sqlJSON(metadata map[string]any) []byte {
	if len(metadata) == 0 {
		return nil
	}
	raw, err := json.Marshal(sanitizeMetadata(metadata))
	if err != nil {
		return nil
	}
	return raw
}

func truncateString(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
