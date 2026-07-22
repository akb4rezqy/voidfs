package config

import (
	"os"
	"strconv"
)

type Config struct {
	Addr           string
	RootDir        string
	SessionSecret  string
	MaxUploadBytes int64
	MaxEditBytes   int64
	AllowedUser    string
}

func Load() Config {
	return Config{
		Addr:           getEnv("APP_ADDR", ":8787"),
		RootDir:        getEnv("APP_ROOT_DIR", "/"),
		SessionSecret:  getEnv("APP_SESSION_SECRET", "change-me"),
		MaxUploadBytes: getEnvInt64("APP_MAX_UPLOAD_BYTES", 10*1024*1024),
		MaxEditBytes:   getEnvInt64("APP_MAX_EDIT_BYTES", 1*1024*1024),
		AllowedUser:    getEnv("APP_ALLOWED_USER", "root"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return parsed
		}
	}
	return fallback
}
