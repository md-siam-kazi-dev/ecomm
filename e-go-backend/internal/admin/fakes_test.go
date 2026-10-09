package admin

import (
	"context"
	"time"
)

// fakeUserRepo records the query it receives and returns canned rows.
type fakeUserRepo struct {
	lastQuery ListUsersQuery
	records   []UserRecord
	total     int64
	err       error
}

func (f *fakeUserRepo) ListUsers(_ context.Context, q ListUsersQuery) ([]UserRecord, int64, error) {
	f.lastQuery = q
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.records, f.total, nil
}

func sampleRecord(id, email, name string, roles []string) UserRecord {
	return UserRecord{
		ID:            id,
		Email:         email,
		Name:          name,
		Status:        "active",
		Roles:         roles,
		EmailVerified: true,
		CreatedAt:     time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}
}
