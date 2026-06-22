package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig
	DB	DBConfig
	JWT JWTConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	Host		 string
	Port		 string
	User		 string
	Password string
	Name		 string
	SSLMode	 string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("JWT_EXPIRY", "24h")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("no .env file found, reading from env: %v", err)
	}

	cfg := &Config{
		App: AppConfig{
			Port: viper.GetString("APP_PORT"),
			Env:  viper.GetString("APP_ENV"),
		},
		DB: DBConfig{
			Host:			viper.GetString("DB_HOST"),
			Port: 		viper.GetString("DB_PORT"),
			User: 		viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:			viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		JWT: JWTConfig{
			Secret: viper.GetString("JWT_SECRET"),
			Expiry: viper.GetDuration("JWT_EXPIRY"),
		},
	}

	if err := cfg.validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	return cfg
}

// devJWTSecret is the placeholder shipped in .env.example; it must never reach
// production.
const devJWTSecret = "dev-only-change-me-to-a-long-random-string"

// validate enforces the invariants the app cannot safely run without. Basic
// requirements apply everywhere; the stricter checks only fire in production so
// local development stays frictionless.
func (c *Config) validate() error {
	var problems []string

	if c.JWT.Secret == "" {
		problems = append(problems, "JWT_SECRET is required")
	}
	if c.DB.Host == "" {
		problems = append(problems, "DB_HOST is required")
	}

	if c.App.Env == "production" {
		if len(c.JWT.Secret) < 32 {
			problems = append(problems, "JWT_SECRET must be at least 32 characters in production")
		}
		if c.JWT.Secret == devJWTSecret {
			problems = append(problems, "JWT_SECRET is still the development placeholder")
		}
		switch c.DB.SSLMode {
		case "require", "verify-ca", "verify-full":
		default:
			problems = append(problems, fmt.Sprintf("DB_SSLMODE must be require/verify-ca/verify-full in production, got %q", c.DB.SSLMode))
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}