package security

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Collector is the single entry point for security events.
//
// It keeps three concerns apart:
//   - Metrics: every request, aggregated in memory, never persisted.
//   - Live stream: every request is pushed to connected admins (bounded queues).
//   - Persistence: only classified/security-relevant events reach PostgreSQL.
//
// That split is what keeps monitoring off the database hot path while still
// giving admins a genuinely live view of traffic.
type Collector struct {
	metrics *Metrics
	writer  *Writer
	broker  *Broker
	now     func() time.Time
}

// NewCollector wires the pieces. Any of metrics/writer/broker may be nil in
// tests, in which case that concern is skipped.
func NewCollector(metrics *Metrics, writer *Writer, broker *Broker) *Collector {
	return &Collector{metrics: metrics, writer: writer, broker: broker, now: time.Now}
}

// RecordRequest handles one HTTP request captured by the monitoring middleware.
// The event's Type is empty when the request was not classified, in which case
// it is counted and streamed but not persisted.
func (c *Collector) RecordRequest(ctx context.Context, ev SecurityEvent) {
	if c == nil {
		return
	}
	ev = c.normalize(ev)
	persist := ev.Type != ""
	ev.Persisted = persist

	if c.metrics != nil {
		c.metrics.Observe(Observation{
			At:         ev.CreatedAt,
			Status:     ev.StatusCode,
			DurationMS: int64(ev.DurationMS),
			Endpoint:   ev.Endpoint,
			IP:         ev.IP,
			Suspicious: ev.IsSuspicious(),
		})
	}
	if c.broker != nil {
		c.broker.Publish(ev)
	}
	if persist {
		c.persist(ctx, ev)
	}
}

// Record handles an explicit, security-relevant event emitted by a service (for
// example a detected refresh-token reuse). It is always persisted and streamed.
//
// It deliberately does not touch the request metrics: those are per-request and
// already counted by RecordRequest, so an explicit event must not inflate them.
func (c *Collector) Record(ctx context.Context, ev SecurityEvent) {
	if c == nil {
		return
	}
	ev = c.normalize(ev)
	if ev.Type == "" {
		ev.Type = EventSuspiciousAPIActivity
	}
	ev.Persisted = true

	if c.broker != nil {
		c.broker.Publish(ev)
	}
	c.persist(ctx, ev)
}

func (c *Collector) normalize(ev SecurityEvent) SecurityEvent {
	ev = ev.WithDefaults(c.clock()())
	if ev.EventID == "" {
		ev.EventID = uuid.NewString()
	}
	if ev.Severity == "" {
		ev.Severity = ev.Type.DefaultSeverity()
	}
	return ev
}

func (c *Collector) persist(ctx context.Context, ev SecurityEvent) {
	if c.writer == nil {
		return
	}
	// Critical events are rare and losing one is unacceptable, so they skip the
	// queue and are written inline.
	if ev.Severity == SeverityCritical {
		_ = c.writer.InsertNow(ctx, ev)
		return
	}
	c.writer.Enqueue(ev)
}

// Stats returns the aggregated overview for a range.
func (c *Collector) Stats(r RangeName) Snapshot {
	if c == nil || c.metrics == nil {
		return Snapshot{Range: r}
	}
	return c.metrics.Snapshot(r)
}

// DroppedEvents reports every place telemetry was shed: the persistence queue
// and the per-subscriber stream queues.
func (c *Collector) DroppedEvents() int64 {
	if c == nil {
		return 0
	}
	var dropped int64
	if c.writer != nil {
		dropped += c.writer.Dropped()
	}
	if c.broker != nil {
		dropped += c.broker.Dropped()
	}
	return dropped
}

func (c *Collector) clock() func() time.Time {
	if c.now == nil {
		return time.Now
	}
	return c.now
}
