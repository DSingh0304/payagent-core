package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	PostgresDSN           string
	RedisURL              string
	MerchantAPIKey        string
	RazorpayKeyID         string
	RazorpaySecret        string
	RazorpayWebhookSecret string
	AgentServiceURL       string
}

func Load() *Config {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	return &Config{
		Port:                  getEnv("MERCHANT_SERVER_PORT", "8080"),
		PostgresDSN:           buildDSN(),
		RedisURL:              getEnv("REDIS_URL", "redis://redis:6379"),
		MerchantAPIKey:        getEnv("MERCHANT_API_KEY", "dev-key"),
		RazorpayKeyID:         getEnv("RAZORPAY_KEY_ID", ""),
		RazorpaySecret:        getEnv("RAZORPAY_KEY_SECRET", ""),
		RazorpayWebhookSecret: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
		AgentServiceURL:       getEnv("AGENT_SERVICE_URL", "http://ai-agent:8001"),
	}
}

func buildDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("POSTGRES_HOST", "postgres"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_USER", "payagent"),
		getEnv("POSTGRES_PASSWORD", "payagent_secret"),
		getEnv("POSTGRES_DB", "payagent"),
	)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
