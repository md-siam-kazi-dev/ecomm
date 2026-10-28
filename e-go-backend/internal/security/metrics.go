package security

import (
	"sort"
	"sync"
	"time"
)

const (
	// minuteRing covers the short ranges (5m/15m/1h) and the last 24h of
	// per-minute resolution.
	minuteRing = 1440
	// hourRing covers 24h/7d at hourly resolution.
	hourRing = 168

	// Bounded cardinality: route patterns are a small closed set, client IPs
	// are not, so both maps are capped and overflow is folded into otherKey.
	maxTrackedEndpoints = 96
	maxTrackedIPs       = 48
	otherKey            = "(other)"

	// topN is how many rows the overview charts show.
	topN = 10

	// maxSeriesPoints caps the volume series so a wide range cannot produce an
	// unusable chart or an expensive snapshot.
	maxSeriesPoints = 400
)

// Observation is the raw, safe metadata captured for every API request. It is
// fed to Metrics for *all* traffic — 2xx included — so the overview numbers are
// totals, not a summary of whatever events happened to be loaded.
type Observation struct {
	At         time.Time
	Status     int
	DurationMS int64
	Endpoint   string
	IP         string
	Suspicious bool
}

// bucketCounts is one aggregation slot.
type bucketCounts struct {
	total       int64
	success     int64
	redirect    int64
	clientErr   int64
	serverErr   int64
	authFailures int64
	rateLimited int64
	suspicious  int64
	latencySumMS int64
	latencyCount int64

	status    map[int]int64
	endpoints map[string]int64
	ips       map[string]int64
}

func (b *bucketCounts) reset() { *b = bucketCounts{} }

func (b *bucketCounts) observe(o Observation) {
	b.total++
	switch {
	case o.Status >= 500:
		b.serverErr++
	case o.Status >= 400:
		b.clientErr++
	case o.Status >= 300:
		b.redirect++
	default:
		b.success++
	}

	switch o.Status {
	case 401:
		b.authFailures++
	case 429:
		b.rateLimited++
	}
	if o.Suspicious {
		b.suspicious++
	}

	if o.DurationMS >= 0 {
		b.latencySumMS += o.DurationMS
		b.latencyCount++
	}

	if b.status == nil {
		b.status = make(map[int]int64, 8)
	}
	b.status[o.Status]++

	if o.Endpoint != "" {
		b.endpoints = bump(b.endpoints, o.Endpoint, maxTrackedEndpoints)
	}
	if o.IP != "" {
		b.ips = bump(b.ips, o.IP, maxTrackedIPs)
	}
}

// bump increments key, folding unknown keys into otherKey once the map is at
// its cardinality limit. Bounded memory is worth approximate long-tail counts.
func bump(m map[string]int64, key string, limit int) map[string]int64 {
	if m == nil {
		m = make(map[string]int64, 8)
	}
	if len(m) >= limit {
		if _, tracked := m[key]; !tracked {
			key = otherKey
		}
	}
	m[key]++
	return m
}

// Metrics aggregates request traffic in bounded rolling buckets. Every update
// is an in-memory increment: there is no database write and no allocation on
// the hot path beyond the first touch of a bucket.
//
// A single mutex guards the buckets. The critical section is a few integer
// increments, and snapshots happen only when an admin loads the overview, so
// contention is far cheaper than the monitoring it replaces.
type Metrics struct {
	mu sync.Mutex

	minuteSlots [minuteRing]int64
	minutes     [minuteRing]bucketCounts

	hourSlots [hourRing]int64
	hours     [hourRing]bucketCounts

	now func() time.Time
}

// NewMetrics builds an empty aggregator.
func NewMetrics() *Metrics {
	return &Metrics{now: time.Now}
}

// Observe records one request.
func (m *Metrics) Observe(o Observation) {
	if o.At.IsZero() {
		o.At = m.clock()()
	}
	minute := o.At.Unix() / 60
	hour := o.At.Unix() / 3600

	m.mu.Lock()
	m.observeInto(m.minutes[:], m.minuteSlots[:], minute, o)
	m.observeInto(m.hours[:], m.hourSlots[:], hour, o)
	m.mu.Unlock()
}

func (m *Metrics) observeInto(buckets []bucketCounts, slots []int64, slot int64, o Observation) {
	if len(buckets) == 0 {
		return
	}
	idx := int(slot % int64(len(buckets)))
	if slots[idx] != slot {
		// A slot from an earlier rotation: start it over.
		buckets[idx].reset()
		slots[idx] = slot
	}
	buckets[idx].observe(o)
}

func (m *Metrics) clock() func() time.Time {
	if m.now == nil {
		return time.Now
	}
	return m.now
}

// StatusCount is one entry of the HTTP status distribution.
type StatusCount struct {
	Status int   `json:"status"`
	Count  int64 `json:"count"`
}

// SeriesPoint is one point of a time series.
type SeriesPoint struct {
	At    time.Time `json:"at"`
	Count int64     `json:"count"`
}

// KeyCount is a ranked endpoint or address.
type KeyCount struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// Snapshot is the aggregated overview for one range.
type Snapshot struct {
	Range             RangeName     `json:"range"`
	From              time.Time     `json:"from"`
	To                time.Time     `json:"to"`
	Total             int64         `json:"total_requests"`
	RequestsPerSecond float64       `json:"requests_per_second"`
	Success           int64         `json:"successful_requests"`
	Redirects         int64         `json:"redirects"`
	ClientErrors      int64         `json:"client_errors"`
	ServerErrors      int64         `json:"server_errors"`
	Failed            int64         `json:"failed_requests"`
	AuthFailures      int64         `json:"failed_auth_attempts"`
	RateLimited       int64         `json:"rate_limit_violations"`
	Suspicious        int64         `json:"suspicious_events"`
	AvgResponseMS     float64       `json:"average_response_ms"`
	Status            []StatusCount `json:"status_distribution"`
	Volume            []SeriesPoint `json:"volume_series"`
	FailedAuth        []SeriesPoint `json:"failed_auth_series"`
	TopEndpoints      []KeyCount    `json:"top_endpoints"`
	TopIPs            []KeyCount    `json:"top_ips"`

	latencySum   int64
	latencyCount int64
}

// Snapshot aggregates the requested window. Short ranges read the minute ring;
// 24h and 7d read the hourly ring, which keeps the merge bounded.
func (m *Metrics) Snapshot(r RangeName) Snapshot {
	if !r.Valid() {
		r = Range15m
	}
	now := m.clock()()
	from := now.Add(-r.Duration())

	useHours := r == Range24h || r == Range7d
	var (
		buckets []bucketCounts
		slots   []int64
		unit    = time.Minute
		step    = 1
	)
	if useHours {
		buckets, slots, unit = m.hours[:], m.hourSlots[:], time.Hour
		if r == Range7d {
			step = 24
		}
	} else if r == Range1h {
		step = 5
	}

	fromSlot := from.UnixNano() / int64(unit)
	toSlot := now.UnixNano() / int64(unit)
	width := int(toSlot-fromSlot) + 1
	if width < 1 {
		width = 1
	}
	points := (width + step - 1) / step
	if points < 1 {
		points = 1
	}
	if points > maxSeriesPoints {
		points = maxSeriesPoints
	}

	volume := make([]SeriesPoint, points)
	failedAuth := make([]SeriesPoint, points)
	for i := 0; i < points; i++ {
		at := time.Unix(0, (fromSlot+int64(i*step))*int64(unit)).UTC()
		volume[i] = SeriesPoint{At: at}
		failedAuth[i] = SeriesPoint{At: at}
	}

	snap := Snapshot{Range: r, From: from.UTC(), To: now.UTC(), Volume: volume, FailedAuth: failedAuth}
	status := make(map[int]int64, 8)
	endpoints := make(map[string]int64, 16)
	ips := make(map[string]int64, 16)

	m.mu.Lock()
	for slot := fromSlot; slot <= toSlot; slot++ {
		if len(buckets) == 0 {
			break
		}
		idx := int(slot % int64(len(buckets)))
		if slots[idx] != slot {
			continue
		}
		b := &buckets[idx]

		snap.Total += b.total
		snap.Success += b.success
		snap.Redirects += b.redirect
		snap.ClientErrors += b.clientErr
		snap.ServerErrors += b.serverErr
		snap.AuthFailures += b.authFailures
		snap.RateLimited += b.rateLimited
		snap.Suspicious += b.suspicious
		snap.latencySum += b.latencySumMS
		snap.latencyCount += b.latencyCount

		for code, n := range b.status {
			status[code] += n
		}
		for key, n := range b.endpoints {
			endpoints[key] += n
		}
		for key, n := range b.ips {
			ips[key] += n
		}

		if group := int((slot - fromSlot) / int64(step)); group >= 0 && group < points {
			volume[group].Count += b.total
			failedAuth[group].Count += b.authFailures
		}
	}
	m.mu.Unlock()

	snap.Failed = snap.ClientErrors + snap.ServerErrors
	if seconds := r.Duration().Seconds(); seconds > 0 {
		snap.RequestsPerSecond = float64(snap.Total) / seconds
	}
	if snap.latencyCount > 0 {
		snap.AvgResponseMS = float64(snap.latencySum) / float64(snap.latencyCount)
	}

	snap.Status = make([]StatusCount, 0, len(status))
	for code, n := range status {
		snap.Status = append(snap.Status, StatusCount{Status: code, Count: n})
	}
	sort.Slice(snap.Status, func(i, j int) bool {
		if snap.Status[i].Count != snap.Status[j].Count {
			return snap.Status[i].Count > snap.Status[j].Count
		}
		return snap.Status[i].Status < snap.Status[j].Status
	})

	snap.TopEndpoints = topKeys(endpoints, topN)
	snap.TopIPs = topKeys(ips, topN)

	return snap
}

func topKeys(m map[string]int64, n int) []KeyCount {
	out := make([]KeyCount, 0, len(m))
	for key, count := range m {
		out = append(out, KeyCount{Key: key, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Key < out[j].Key
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}
