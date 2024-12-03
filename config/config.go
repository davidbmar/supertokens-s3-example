// config/config.go

package config

import (
    "os"
    "github.com/joho/godotenv"
)

// Config is similar to a Python dataclass
type Config struct {
    ServerPort     string
    AWSRegion      string
    AWSBucket      string
    SuperTokensURL string
    SuperTokensKey string
}

func Load() (*Config, error) {
    // Load .env file (similar to python-dotenv)
    if err := godotenv.Load(); err != nil {
        return nil, err
    }

    // In Go, we handle missing env vars explicitly
    // unlike Python's os.getenv() which returns None
    return &Config{
        ServerPort:     getEnvOrDefault("SERVER_PORT", "8080"),
        AWSRegion:      os.Getenv("AWS_REGION"),
        AWSBucket:      os.Getenv("AWS_BUCKET"),
        SuperTokensURL: os.Getenv("SUPERTOKENS_URL"),
        SuperTokensKey: os.Getenv("SUPERTOKENS_KEY"),
    }, nil
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
