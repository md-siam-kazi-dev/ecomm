package config

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Auth     AuthConfig
	Images   ImageConfig
	Payment  PaymentConfig
	Order    OrderConfig
	Log      LogConfig
}

// OrderConfig holds storefront order pricing (flat rates, no per-product tax).
type OrderConfig struct {
	StandardShipping      float64
	ExpressShipping       float64
	FreeShippingThreshold float64
}

// PaymentConfig configures the application-fee workflow. No real provider is
// integrated yet; ApplicationFeeEnabled gates whether a store may activate
// before a verified payment.
type PaymentConfig struct {
	// Provider is the payment gateway name. Empty means "not configured".
	Provider      string
	ProviderKey   string
	WebhookSecret string
	// ApplicationFeeEnabled requires a verified application-fee payment before
	// a store can activate.
	ApplicationFeeEnabled bool
	// ApplicationFeeAmount is a decimal string (never a float), e.g. "1500.00".
	ApplicationFeeAmount string
	// ApplicationFeeCurrency is an ISO-4217 code, e.g. "BDT".
	ApplicationFeeCurrency string
	// Currency is the default order currency, e.g. "usd".
	Currency string
}

// ImageConfig configures third-party image hosting (profile pictures).
type ImageConfig struct {
	// ImgBBAPIKey is required to upload to imgbb.com. When empty, avatar
	// uploads return a clear "not configured" error.
	ImgBBAPIKey string
}

// AuthConfig controls browser-facing authentication behaviour. The JWT secrets
// and token TTLs live in JWTConfig; this covers cookies and the frontend origin.
type AuthConfig struct {
	// CookieSecure sets the Secure flag on auth cookies. Keep false for plain
	// HTTP local development, true (HTTPS) in production.
	CookieSecure bool
	// CookieDomain optionally scopes cookies to a parent domain.
	CookieDomain string
	// FrontendURL is the browser origin used for CORS and for links in
	// verification / password-reset emails.
	FrontendURL string
}

type AppConfig struct {
	Env                 string
	Name                string
	Port                string
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	ShutdownTimeout     time.Duration
	MaxRequestBodyBytes int64
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

type RedisConfig struct {
	URL          string
	PoolSize     int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	Issuer        string
	Audience      string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type LogConfig struct {
	Level string
}

// Load reads configuration from environment variables, applying defaults for
// non-secret values and returning an error when required secrets are missing.
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Env:                 getEnv("APP_ENV", "development"),
			Name:                getEnv("APP_NAME", "ShopKart"),
			Port:                getEnv("APP_PORT", "8080"),
			ReadTimeout:         getEnvDuration("APP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:        getEnvDuration("APP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:         getEnvDuration("APP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:     getEnvDuration("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
			MaxRequestBodyBytes: getEnvInt64("APP_MAX_REQUEST_BODY_BYTES", 1<<20),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			MaxConns:        int32(getEnvInt("DB_MAX_CONNS", 20)),
			MinConns:        int32(getEnvInt("DB_MIN_CONNS", 2)),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			ConnectTimeout:  getEnvDuration("DB_CONNECT_TIMEOUT", 10*time.Second),
		},
		Redis: RedisConfig{
			URL:          getEnv("REDIS_URL", ""),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			DialTimeout:  getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
			Issuer:        getEnv("JWT_ISSUER", "e-go-backend"),
			Audience:      getEnv("JWT_AUDIENCE", "e-go-api"),
			AccessTTL:     getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    getEnvDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			AllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Authorization", "Content-Type", "Idempotency-Key", "X-Request-Id"}),
		},
		Auth: AuthConfig{
			CookieSecure: getEnvBool("COOKIE_SECURE", false),
			CookieDomain: getEnv("COOKIE_DOMAIN", ""),
			FrontendURL:  getEnv("FRONTEND_URL", ""),
		},
		Images: ImageConfig{
			ImgBBAPIKey: getEnv("IMGBB_API_KEY", ""),
		},
		Payment: PaymentConfig{
			Provider:               getEnv("PAYMENT_PROVIDER", ""),
			ProviderKey:            getEnv("PAYMENT_PROVIDER_KEY", ""),
			WebhookSecret:          getEnv("PAYMENT_WEBHOOK_SECRET", ""),
			ApplicationFeeEnabled:  getEnvBool("APPLICATION_FEE_ENABLED", false),
			ApplicationFeeAmount:   getEnv("APPLICATION_FEE_AMOUNT", "0.00"),
			ApplicationFeeCurrency: getEnv("APPLICATION_FEE_CURRENCY", "BDT"),
			Currency:               getEnv("PAYMENT_CURRENCY", "usd"),
		},
		Order: OrderConfig{
			StandardShipping:      getEnvFloat("ORDER_STANDARD_SHIPPING", 5.99),
			ExpressShipping:       getEnvFloat("ORDER_EXPRESS_SHIPPING", 12.99),
			FreeShippingThreshold: getEnvFloat("ORDER_FREE_SHIPPING_THRESHOLD", 75),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	// The frontend origin must be an allowed CORS origin so cookie-authenticated
	// requests from the storefront are accepted.
	if cfg.Auth.FrontendURL != "" && !slices.Contains(cfg.CORS.AllowedOrigins, cfg.Auth.FrontendURL) {
		cfg.CORS.AllowedOrigins = append(cfg.CORS.AllowedOrigins, cfg.Auth.FrontendURL)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Env, "production")
}

func (c *Config) validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("config: DATABASE_URL is required")
	}
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("config: JWT_ACCESS_SECRET is required")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("config: JWT_REFRESH_SECRET is required")
	}
	if c.IsProduction() {
		if len(c.JWT.AccessSecret) < 32 || len(c.JWT.RefreshSecret) < 32 {
			return fmt.Errorf("config: JWT secrets must be at least 32 characters in production")
		}
	}
	if c.Payment.ApplicationFeeEnabled {
		if strings.TrimSpace(c.Payment.ApplicationFeeAmount) == "" {
			return fmt.Errorf("config: APPLICATION_FEE_AMOUNT is required when the application fee is enabled")
		}
		if len(c.Payment.ApplicationFeeCurrency) != 3 {
			return fmt.Errorf("config: APPLICATION_FEE_CURRENCY must be a 3-letter code")
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if parsed, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
