package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Host                 string
	Port                 int
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	ShutdownTimeout      time.Duration
	SQLitePath           string
	MigrationsDir        string
	OutboundTimeout      time.Duration
	OutboundRetries      int
	OutboundBackoff      time.Duration
	PGHEndpoint          string
	PGHPollInterval      time.Duration
	TickTickAPIRoot      string
	TickTickAuthURL      string
	TickTickTokenURL     string
	TickTickClientID     string
	TickTickClientSecret string
	TickTickRedirectURI  string
	TickTickToken        string
	TickTickProject      string
	TickTickPoll         time.Duration
	AllowedOrigins       []string
}

func Load() (Config, error) {
	_ = godotenv.Load(".env")

	port, err := intFromEnv("APP_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	readTimeout, err := durationFromEnv("APP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	writeTimeout, err := durationFromEnv("APP_WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := durationFromEnv("APP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	outboundTimeout, err := durationFromEnv("OUTBOUND_HTTP_TIMEOUT", 8*time.Second)
	if err != nil {
		return Config{}, err
	}

	outboundRetries, err := intFromEnv("OUTBOUND_RETRY_COUNT", 2)
	if err != nil {
		return Config{}, err
	}

	outboundBackoff, err := durationFromEnv("OUTBOUND_RETRY_BACKOFF", 300*time.Millisecond)
	if err != nil {
		return Config{}, err
	}

	pghPollInterval, err := durationFromEnv("PGHST_POLL_INTERVAL", 12*time.Hour)
	if err != nil {
		return Config{}, err
	}

	tickTickPoll, err := durationFromEnv("TICKTICK_POLL_INTERVAL", 10*time.Minute)
	if err != nil {
		return Config{}, err
	}

	sqlitePath := stringFromEnv("SQLITE_PATH", "./hoel.db")
	if sqlitePath == "" {
		return Config{}, fmt.Errorf("SQLITE_PATH cannot be empty")
	}

	migrationsDir := stringFromEnv("MIGRATIONS_DIR", "./migrations")
	if migrationsDir == "" {
		return Config{}, fmt.Errorf("MIGRATIONS_DIR cannot be empty")
	}

	return Config{
		Host:                 stringFromEnv("APP_HOST", "127.0.0.1"),
		Port:                 port,
		ReadTimeout:          readTimeout,
		WriteTimeout:         writeTimeout,
		ShutdownTimeout:      shutdownTimeout,
		SQLitePath:           sqlitePath,
		MigrationsDir:        migrationsDir,
		OutboundTimeout:      outboundTimeout,
		OutboundRetries:      outboundRetries,
		OutboundBackoff:      outboundBackoff,
		PGHEndpoint:          stringFromEnv("PGHST_ENDPOINT", ""),
		PGHPollInterval:      pghPollInterval,
		TickTickAPIRoot:      stringFromEnv("TICKTICK_API_ROOT", "https://api.ticktick.com/open/v1"),
		TickTickAuthURL:      stringFromEnv("TICKTICK_AUTH_URL", "https://ticktick.com/oauth/authorize"),
		TickTickTokenURL:     stringFromEnv("TICKTICK_TOKEN_URL", "https://ticktick.com/oauth/token"),
		TickTickClientID:     stringFromEnv("TICKTICK_CLIENT_ID", ""),
		TickTickClientSecret: stringFromEnv("TICKTICK_CLIENT_SECRET", ""),
		TickTickRedirectURI:  stringFromEnv("TICKTICK_REDIRECT_URI", ""),
		TickTickToken:        stringFromEnv("TICKTICK_ACCESS_TOKEN", ""),
		TickTickProject:      stringFromEnv("TICKTICK_PROJECT_ID", ""),
		TickTickPoll:         tickTickPoll,
		AllowedOrigins: listFromEnv(
			"APP_ALLOWED_ORIGINS",
			[]string{"http://localhost:5173", "http://127.0.0.1:5173"},
		),
	}, nil
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func stringFromEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func intFromEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return parsed, nil
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return parsed, nil
}

func listFromEnv(key string, fallback []string) []string {
	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		return append([]string{}, fallback...)
	}

	segments := strings.Split(value, ",")
	values := make([]string, 0, len(segments))
	for _, segment := range segments {
		trimmed := strings.TrimSpace(segment)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return append([]string{}, fallback...)
	}

	return values
}
