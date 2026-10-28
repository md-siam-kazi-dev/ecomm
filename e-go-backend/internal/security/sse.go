package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/md-siam-kazi/e-go-backend/internal/auth"
	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
	"github.com/md-siam-kazi/e-go-backend/pkg/response"
)

// Streamer serves the authenticated Server-Sent Events feed.
//
// Authentication is the existing HttpOnly access-token cookie: the browser's
// EventSource cannot set an Authorization header, but it does send same-site
// cookies automatically, and the frontend reaches this endpoint through its own
// origin (`/api/v1/...` rewrite). No token is ever placed in the query string.
type Streamer struct {
	broker    *Broker
	heartbeat time.Duration
	subBuffer int
}

// NewStreamer builds a streamer.
func NewStreamer(broker *Broker, heartbeat time.Duration, subBuffer int) *Streamer {
	if heartbeat <= 0 {
		heartbeat = 20 * time.Second
	}
	if subBuffer <= 0 {
		subBuffer = 64
	}
	return &Streamer{broker: broker, heartbeat: heartbeat, subBuffer: subBuffer}
}

// Serve streams events until the client disconnects, the broker closes, or the
// access token expires.
func (s *Streamer) Serve(w http.ResponseWriter, r *http.Request) {
	if s == nil || s.broker == nil {
		response.Error(w, r, apperr.ErrInternal)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, r, apperr.ErrInternal)
		return
	}

	// Tell the monitoring middleware this request became a long-lived stream, so
	// its minutes-long duration does not distort the average response time.
	MarkStreaming(r.Context())

	// The server's global WriteTimeout would kill the stream after 15s; clear the
	// deadline for this connection only.
	if controller := http.NewResponseController(w); controller != nil {
		_ = controller.SetWriteDeadline(time.Time{})
	}

	header := w.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache, no-transform")
	// Defensive: tells nginx-style proxies not to buffer the stream.
	header.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sub := s.broker.Subscribe(s.subBuffer)
	defer s.broker.Unsubscribe(sub)

	// Resume from where the client left off. A miss (id outside the replay
	// window) is not an error: the client re-syncs from the REST history.
	if lastID := strings.TrimSpace(r.Header.Get("Last-Event-ID")); lastID != "" {
		for _, ev := range s.broker.Replay(lastID) {
			if !writeEvent(w, flusher, ev) {
				return
			}
		}
	}
	if !writeComment(w, flusher, "connected") {
		return
	}

	ticker := time.NewTicker(s.heartbeat)
	defer ticker.Stop()

	expiry := tokenExpiryTimer(r)

	var expiryC <-chan time.Time
	if expiry != nil {
		defer expiry.Stop()
		expiryC = expiry.C
	}

	for {
		select {
		case <-r.Context().Done():
			// Client went away.
			return
		case <-expiryC:
			// The access token is expiring. End the stream cleanly; EventSource
			// reconnects automatically and the REST client refreshes the cookie.
			writeComment(w, flusher, "reconnect: access token expired")
			return
		case <-ticker.C:
			if !writeComment(w, flusher, "ping") {
				return
			}
		case ev, open := <-sub.Events:
			if !open {
				// Broker shut down.
				return
			}
			if !writeEvent(w, flusher, ev) {
				return
			}
		}
	}
}

// tokenExpiryTimer closes the stream shortly before the access token expires,
// so an open stream can never outlive the credential that authorised it.
func tokenExpiryTimer(r *http.Request) *time.Timer {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok || identity.ExpiresAt <= 0 {
		return nil
	}
	remaining := time.Until(time.Unix(identity.ExpiresAt, 0)) - 5*time.Second
	if remaining < 0 {
		remaining = 0
	}
	return time.NewTimer(remaining)
}

func writeEvent(w http.ResponseWriter, flusher http.Flusher, ev SecurityEvent) bool {
	payload, err := json.Marshal(ToStreamEvent(ev))
	if err != nil {
		// Skip an unencodable event rather than tearing down the stream.
		return true
	}
	// The JSON encoder escapes newlines, so the data field stays single-line.
	if _, err := fmt.Fprintf(w, "id: %s\nevent: security\ndata: %s\n\n", ev.EventID, payload); err != nil {
		return false
	}
	flusher.Flush()
	return true
}

// writeComment sends an SSE comment, which keeps proxies from idling the
// connection out without delivering a spurious event to the client.
func writeComment(w http.ResponseWriter, flusher http.Flusher, text string) bool {
	if _, err := fmt.Fprintf(w, ": %s\n\n", text); err != nil {
		return false
	}
	flusher.Flush()
	return true
}
