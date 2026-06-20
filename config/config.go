package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig
	DB	DBConfig
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

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

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
	}
}