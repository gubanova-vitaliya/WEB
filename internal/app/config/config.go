package config

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type JWTConfig struct {
	Secret        string
	ExpiresIn     time.Duration
	SigningMethod jwt.SigningMethod
}

type Config struct {
	ServiceHost string
	ServicePort int
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")

	// Устанавливаем значения по умолчанию
	viper.SetDefault("ServiceHost", "localhost")
	viper.SetDefault("ServicePort", 8080)

	err := viper.ReadInConfig()
	if err != nil {
		log.Warnf("Config file not found, using defaults: %v", err)
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	log.Info("Configuration loaded successfully")
	return cfg, nil
}

func getSigningMethod(algorithm string) jwt.SigningMethod {
	switch algorithm {
	case "HS256":
		return jwt.SigningMethodHS256
	case "HS384":
		return jwt.SigningMethodHS384
	case "HS512":
		return jwt.SigningMethodHS512
	default:
		return jwt.SigningMethodHS256
	}
}
