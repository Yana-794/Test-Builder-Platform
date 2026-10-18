package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

// Config — все настройки приложения, собранные из переменных окружения.
// Читается один раз при старте, дальше передаётся по коду.
type Config struct {
	AppEnv         string // "development" | "production"
	HTTPAddr       string // "0.0.0.0:8080"
	FrontendOrigin string // "http://localhost" — для CORS
	DatabaseURL    string 
}

// Load читает env, валидирует и возвращает Config.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		HTTPAddr:       getEnv("HTTP_ADDR", "0.0.0.0:8080"),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost"),
		DatabaseURL:    strings.TrimSpace(os.Getenv("DATABASE_URL")),
	}

	cfg.AppEnv = strings.ToLower(strings.TrimSpace(cfg.AppEnv))
	cfg.FrontendOrigin = strings.TrimRight(strings.TrimSpace(cfg.FrontendOrigin), "/")

	if cfg.AppEnv != "development" && cfg.AppEnv != "production" {
		return Config{}, errors.New("APP_ENV must be development or production")
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if strings.TrimSpace(cfg.HTTPAddr) == "" {
		return Config{}, errors.New("HTTP_ADDR is required")
	}

	if err := validateOrigin(cfg.FrontendOrigin); err != nil {
		return Config{}, fmt.Errorf("FRONTEND_ORIGIN: %w", err)
	}

	dbURL, err := url.Parse(cfg.DatabaseURL)
	if err != nil ||
		(dbURL.Scheme != "postgres" && dbURL.Scheme != "postgresql") ||
		dbURL.Host == "" {
		return Config{}, errors.New("DATABASE_URL must be a valid PostgreSQL URL")
	}

	sslmode := dbURL.Query().Get("sslmode")

	switch cfg.AppEnv {
	case "production":

		if sslmode != "verify-full" {
			return Config{}, errors.New(
				"production DATABASE_URL must use sslmode=verify-full",
			)
		}
	case "development":
		
		if sslmode == "" {
			cfg.DatabaseURL = cfg.DatabaseURL + "?sslmode=disable"
		}
	}

	return cfg, nil
}

// MustLoad — как Load, но падает с log.Fatal при ошибке.
func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	return cfg
}



func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func validateOrigin(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Host == "" {
		return errors.New("host is required")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("scheme must be http or https")
	}
	if u.Path != "" && u.Path != "/" {
		return errors.New("path must be empty")
	}
	if u.RawQuery != "" {
		return errors.New("query must be empty")
	}
	if u.Fragment != "" {
		return errors.New("fragment must be empty")
	}
	return nil
}