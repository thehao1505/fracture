package config

import (
	"log"
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

	return &Config{
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
}