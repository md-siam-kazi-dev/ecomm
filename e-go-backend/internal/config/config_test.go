package config

import (
	"testing"
	"time"
)

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_ACCESS_SECRET", "access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-secret")

	if _, err := Load(); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoad_RequiresJWTSecrets(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
	t.Setenv("JWT_ACCESS_SECRET", "")
	t.Setenv("JWT_REFRESH_SECRET", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected error when JWT secrets are missing")
	}
}

func TestLoad_AppliesDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
	t.Setenv("JWT_ACCESS_SECRET", "access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-secret")
	t.Setenv("APP_PORT", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("JWT_ACCESS_TTL", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.App.Port != "8080" {
		t.Errorf("expected default port 8080, got %q", cfg.App.Port)
	}
	if cfg.App.Env != "development" {
		t.Errorf("expected default env development, got %q", cfg.App.Env)
	}
	if cfg.JWT.AccessTTL != 15*time.Minute {
		t.Errorf("expected default access TTL 15m, got %s", cfg.JWT.AccessTTL)
	}
	if len(cfg.CORS.AllowedOrigins) == 0 {
		t.Error("expected default CORS origins to be non-empty")
	}
	if cfg.IsProduction() {
		t.Error("development env should not report as production")
	}
}

func TestLoad_ParsesOverrides(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
	t.Setenv("JWT_ACCESS_SECRET", "access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-secret")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("DB_MAX_CONNS", "42")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.com, https://b.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.App.Port != "9090" {
		t.Errorf("expected port 9090, got %q", cfg.App.Port)
	}
	if cfg.Database.MaxConns != 42 {
		t.Errorf("expected max conns 42, got %d", cfg.Database.MaxConns)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Fatalf("expected 2 CORS origins, got %d", len(cfg.CORS.AllowedOrigins))
	}
	if cfg.CORS.AllowedOrigins[1] != "https://b.com" {
		t.Errorf("expected trimmed origin, got %q", cfg.CORS.AllowedOrigins[1])
	}
}
