package audit

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type fakeQuerier struct {
	calls int
	sql   string
	args  []any
	err   error
}

func (f *fakeQuerier) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	f.calls++
	f.sql = sql
	f.args = args
	return pgconn.CommandTag{}, f.err
}

func TestRecordDefaultsAndContext(t *testing.T) {
	q := &fakeQuerier{}
	ctx := context.WithValue(context.Background(), requestInfoKey, RequestInfo{
		IP: "203.0.113.7", UserAgent: "test-agent", RequestID: "req-1",
	})

	actor := "11111111-1111-1111-1111-111111111111"
	entity := "22222222-2222-2222-2222-222222222222"
	if err := Record(ctx, q, Event{
		ActorID:    &actor,
		ActorRole:  RoleAdmin,
		Action:     ActionAdminProductApproved,
		EntityType: EntityProduct,
		EntityID:   &entity,
		NewValues:  map[string]any{"status": "active", "password": "secret"},
	}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	if q.calls != 1 {
		t.Fatalf("expected 1 exec, got %d", q.calls)
	}
	// args: user, role, action, entityType, entityID, old, new, result, reason, ip, ua, requestID, metadata
	if got := q.args[1]; got != RoleAdmin {
		t.Errorf("actor_role = %v, want %v", got, RoleAdmin)
	}
	if got := q.args[7]; got != ResultSuccess {
		t.Errorf("result = %v, want %v (default)", got, ResultSuccess)
	}
	if got := q.args[9]; got == nil || got.(string) != "203.0.113.7" {
		t.Errorf("ip not taken from context: %v", got)
	}
	if got := q.args[10]; got != "test-agent" {
		t.Errorf("user agent not taken from context: %v", got)
	}
	if got := q.args[11]; got != "req-1" {
		t.Errorf("request id not taken from context: %v", got)
	}
	newJSON := string(q.args[6].([]byte))
	if !strings.Contains(newJSON, redacted) {
		t.Errorf("sensitive value in new_values was not redacted: %s", newJSON)
	}
}

func TestRecordInvalidEntityFallsBackToMetadata(t *testing.T) {
	q := &fakeQuerier{}
	entity := "not-a-uuid"
	if err := Record(context.Background(), q, Event{
		Action:     "x",
		EntityType: EntityProduct,
		EntityID:   &entity,
	}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if q.args[4] != nil {
		t.Errorf("invalid entity id should store NULL, got %v", q.args[4])
	}
	meta := string(q.args[12].([]byte))
	if !strings.Contains(meta, "not-a-uuid") {
		t.Errorf("entity_ref not preserved in metadata: %s", meta)
	}
}

func TestRecordInvalidIPStoredAsNull(t *testing.T) {
	q := &fakeQuerier{}
	if err := Record(context.Background(), q, Event{Action: "x", IP: "not-an-ip"}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if q.args[9] != nil {
		t.Errorf("invalid ip should store NULL, got %v", q.args[9])
	}
}

func TestRecordNilQuerierNoop(t *testing.T) {
	if err := Record(context.Background(), nil, Event{Action: "x"}); err != nil {
		t.Fatalf("nil querier should be a no-op: %v", err)
	}
}

func TestRecorderDisabledAndError(t *testing.T) {
	var nilRec *Recorder
	if nilRec.Enabled() {
		t.Errorf("nil recorder should be disabled")
	}
	if err := nilRec.Log(context.Background(), Event{Action: "x"}); err != nil {
		t.Errorf("nil recorder Log should be a no-op: %v", err)
	}
	if NewRecorder(nil).Enabled() {
		t.Errorf("recorder with nil pool should be disabled")
	}
}

func TestRecordPropagatesExecError(t *testing.T) {
	sentinel := errors.New("boom")
	q := &fakeQuerier{err: sentinel}
	if err := Record(context.Background(), q, Event{Action: "x"}); !errors.Is(err, sentinel) {
		t.Errorf("expected exec error to propagate, got %v", err)
	}
}

func TestLabelAndActions(t *testing.T) {
	if got := Label(ActionAuthLoginFailed); got != "Failed sign-in" {
		t.Errorf("label = %q", got)
	}
	if got := Label("vendor.unknown_thing"); got != "Vendor unknown thing" {
		t.Errorf("fallback label = %q", got)
	}
	actions := Actions()
	if len(actions) == 0 {
		t.Fatalf("expected a non-empty action catalogue")
	}
	for _, a := range actions {
		if a.Value == "" || a.Label == "" {
			t.Errorf("incomplete action entry: %+v", a)
		}
	}
}

func TestNormalizeDefaultsAndAllowLists(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got := normalize(ListQuery{
		Page:       -3,
		Limit:      7,
		ActorRole:  "ADMIN",
		Result:     "SUCCESS",
		EntityType: "PRODUCT",
		Search:     "  hello  ",
		From:       &from,
	})
	if got.Page != 1 || got.Limit != defaultLimit {
		t.Errorf("page/limit not defaulted: %+v", got)
	}
	if got.ActorRole != RoleAdmin || got.Result != ResultSuccess || got.EntityType != EntityProduct {
		t.Errorf("allow-list normalisation failed: %+v", got)
	}
	if got.Search != "hello" {
		t.Errorf("search not trimmed: %q", got.Search)
	}
}

func TestNormalizeRejectsUnknownFilters(t *testing.T) {
	got := normalize(ListQuery{ActorRole: "hacker", Result: "maybe", EntityType: "unicorn"})
	if got.ActorRole != "" || got.Result != "" || got.EntityType != "" {
		t.Errorf("unknown filters should be dropped: %+v", got)
	}
}

func TestBoundRangeClamps(t *testing.T) {
	// Only an upper bound: a lower bound is derived, within the max window.
	to := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	from, gotTo := boundRange(nil, &to)
	if from == nil || gotTo == nil {
		t.Fatalf("expected both bounds, got %v/%v", from, gotTo)
	}
	if gotTo.Sub(*from) > time.Duration(maxRangeDays)*24*time.Hour {
		t.Errorf("range not clamped: %v", gotTo.Sub(*from))
	}

	// Over-long explicit window is clamped.
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f, tt := boundRange(&start, &end)
	if tt.Sub(*f) > time.Duration(maxRangeDays)*24*time.Hour {
		t.Errorf("over-long range not clamped: %v", tt.Sub(*f))
	}
}

func TestCategory(t *testing.T) {
	failure := ResultFailure
	success := ResultSuccess
	cases := []struct {
		action string
		result *string
		want   string
	}{
		{ActionAuthLoginSuccess, &success, "security"},
		{ActionAuthLoginFailed, &failure, "security"},
		{ActionAdminProductApproved, &success, "admin"},
		{ActionVendorProductUpdated, &success, "vendor"},
		{"mystery.event", &success, "system"},
	}
	for _, c := range cases {
		if got := category(c.action, c.result); got != c.want {
			t.Errorf("category(%q) = %q, want %q", c.action, got, c.want)
		}
	}
}
