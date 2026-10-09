package admin

import (
	"context"
	"strings"
	"testing"
)

func TestListUsers_NormalizesQuery(t *testing.T) {
	cases := []struct {
		name       string
		in         ListUsersQuery
		wantPage   int
		wantLimit  int
		wantRole   string
		wantSearch string
	}{
		{"defaults", ListUsersQuery{}, 1, defaultLimit, "", ""},
		{"negative page", ListUsersQuery{Page: -3}, 1, defaultLimit, "", ""},
		{"invalid limit falls back", ListUsersQuery{Limit: 7}, 1, defaultLimit, "", ""},
		{"allowed limit kept", ListUsersQuery{Limit: 50}, 1, 50, "", ""},
		{"customer alias", ListUsersQuery{Role: "customer"}, 1, defaultLimit, "user", ""},
		{"valid role", ListUsersQuery{Role: "vendor"}, 1, defaultLimit, "vendor", ""},
		{"unknown role dropped", ListUsersQuery{Role: "nonsense"}, 1, defaultLimit, "", ""},
		{"search trimmed", ListUsersQuery{Search: "  siam  "}, 1, defaultLimit, "", "siam"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := &fakeUserRepo{}
			_, _, err := NewService(repo).ListUsers(context.Background(), c.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.lastQuery.Page != c.wantPage || repo.lastQuery.Limit != c.wantLimit {
				t.Errorf("page/limit = %d/%d, want %d/%d",
					repo.lastQuery.Page, repo.lastQuery.Limit, c.wantPage, c.wantLimit)
			}
			if repo.lastQuery.Role != c.wantRole {
				t.Errorf("role = %q, want %q", repo.lastQuery.Role, c.wantRole)
			}
			if repo.lastQuery.Search != c.wantSearch {
				t.Errorf("search = %q, want %q", repo.lastQuery.Search, c.wantSearch)
			}
		})
	}
}

func TestListUsers_TruncatesLongSearch(t *testing.T) {
	repo := &fakeUserRepo{}
	long := strings.Repeat("a", maxSearchLen+50)

	if _, _, err := NewService(repo).ListUsers(context.Background(), ListUsersQuery{Search: long}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.lastQuery.Search) != maxSearchLen {
		t.Errorf("search length = %d, want %d", len(repo.lastQuery.Search), maxSearchLen)
	}
}

func TestListUsers_MapsEffectiveRole(t *testing.T) {
	repo := &fakeUserRepo{
		records: []UserRecord{
			sampleRecord("u-1", "a@x.com", "Alpha", []string{"user", "admin"}),
			sampleRecord("u-2", "b@x.com", "Beta", []string{"user"}),
		},
		total: 2,
	}

	items, pagination, err := NewService(repo).ListUsers(context.Background(), ListUsersQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items[0].Role != "admin" {
		t.Errorf("expected highest role admin, got %q", items[0].Role)
	}
	if items[1].Role != "user" {
		t.Errorf("expected role user, got %q", items[1].Role)
	}
	if pagination.Total != 2 {
		t.Errorf("total = %d, want 2", pagination.Total)
	}
}

func TestPrimaryRole(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"user"}, "user"},
		{[]string{"user", "vendor"}, "vendor"},
		{[]string{"vendor", "admin"}, "admin"},
		{nil, "user"},
	}
	for _, c := range cases {
		if got := primaryRole(c.in); got != c.want {
			t.Errorf("primaryRole(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
