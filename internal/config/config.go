package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr            string
	NodeID              string
	JWTSecret           string
	TokenTTL            time.Duration
	MySQLDSN            string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int
	EnableRabbitMQ      bool
	RabbitMQURL         string
	EnableMinIO         bool
	MinIOEndpoint       string
	MinIOAccessKey      string
	MinIOSecretKey      string
	MinIOBucket         string
	MinIOUseSSL         bool
	MinIOPublicBaseURL  string
	EnableElasticsearch bool
	ElasticsearchURL    string
	ElasticsearchIndex  string
}

func Load() Config {
	return Config{
		HTTPAddr:            getenv("IM_ADDR", ":8080"),
		NodeID:              getenv("IM_NODE_ID", "node-1"),
		JWTSecret:           getenv("IM_JWT_SECRET", "change-me"),
		TokenTTL:            time.Duration(getenvInt("IM_TOKEN_TTL_HOURS", 168)) * time.Hour,
		MySQLDSN:            getenv("IM_MYSQL_DSN", "root:123456@tcp(127.0.0.1:3306)/im_chat?parseTime=true&charset=utf8mb4&loc=Local"),
		RedisAddr:           getenv("IM_REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:       os.Getenv("IM_REDIS_PASSWORD"),
		RedisDB:             getenvInt("IM_REDIS_DB", 0),
		EnableRabbitMQ:      getenvBool("IM_ENABLE_RABBITMQ", true),
		RabbitMQURL:         getenv("IM_RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/"),
		EnableMinIO:         getenvBool("IM_ENABLE_MINIO", true),
		MinIOEndpoint:       getenv("IM_MINIO_ENDPOINT", "127.0.0.1:9000"),
		MinIOAccessKey:      getenv("IM_MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:      getenv("IM_MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:         getenv("IM_MINIO_BUCKET", "im-chat"),
		MinIOUseSSL:         getenvBool("IM_MINIO_USE_SSL", false),
		MinIOPublicBaseURL:  getenv("IM_MINIO_PUBLIC_BASE_URL", "http://127.0.0.1:9000"),
		EnableElasticsearch: getenvBool("IM_ENABLE_ELASTICSEARCH", true),
		ElasticsearchURL:    getenv("IM_ELASTICSEARCH_URL", "http://127.0.0.1:9200"),
		ElasticsearchIndex:  getenv("IM_ELASTICSEARCH_INDEX", "messages"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "False", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}
