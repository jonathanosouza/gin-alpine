// Package configs
package configs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gin-alpine/src/pkg/utils"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

type Config struct {
	AppName          string
	Env              string
	Port             string
	RedisURL         string
	AvailableAPIKeys []string
	CSRFSecret       string
	DBConn           string
	EmailSender      string
	EmailSenderPass  string
	EmailSMTP        string
	SMTPAddress      string
	AdminEmail       string
	AdminPass        string
}

func LoadConfig() (*Config, error) {
	env := utils.GetEnvDefault("ENV", "development")
	if env == "development" {
		_ = godotenv.Load(".env.dev")
		_ = godotenv.Load(filepath.Join("..", ".env.dev"))
		_ = godotenv.Load(".env")
		_ = godotenv.Load(filepath.Join("..", ".env"))
	}
	if env == "test" {
		_ = godotenv.Load(".env.test")
		_ = godotenv.Load(filepath.Join("..", ".env.test"))
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		return nil, errors.New("HTTP_PORT is required")
	}
	appName := utils.GetEnvDefault("APP_NAME", "go_templ")
	availableAPIKeysStr := os.Getenv("AVAILABLE_API_KEYS")
	if availableAPIKeysStr == "" {
		return nil, errors.New("AVAILABLE_API_KEYS is required")
	}
	keys := strings.Split(availableAPIKeysStr, ",")
	for i := range keys {
		keys[i] = strings.TrimSpace(keys[i])
	}
	csrf := utils.GetEnvDefault("CSRF_SECRET", "e7962210f5b7a175")
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, errors.New("REDIS_URL is required")
	}

	adminEmail := utils.GetEnvDefault("ADMIN_EMAIL", "admin@email.com")
	adminPass := utils.GetEnvDefault("ADMIN_PASS", "pb_admin")

	dbConn := os.Getenv("GOOSE_DBSTRING")
	if dbConn == "" {
		return nil, errors.New("GOOSE_DBSTRING is required")
	}

	emailSender := os.Getenv("EMAIL_SENDER")
	if emailSender == "" {
		return nil, errors.New("EMAIL_SENDER is required")
	}

	emailSenderPass := os.Getenv("EMAIL_SENDER_PASS")
	if emailSenderPass == "" {
		return nil, errors.New("EMAIL_SENDER_PASS is required")
	}

	emailSMTP := os.Getenv("EMAIL_SMTP")
	if emailSMTP == "" {
		return nil, errors.New("EMAIL_SMTP is required")
	}

	smtpAddress := os.Getenv("SMTP_ADDRESS")
	if smtpAddress == "" {
		return nil, errors.New("SMTP_ADDRESS is required")
	}

	return &Config{
		AppName:          appName,
		Env:              env,
		Port:             port,
		AvailableAPIKeys: keys,
		CSRFSecret:       csrf,
		RedisURL:         redisURL,
		DBConn:           dbConn,
		EmailSender:      emailSender,
		EmailSenderPass:  emailSenderPass,
		EmailSMTP:        emailSMTP,
		SMTPAddress:      smtpAddress,
		AdminEmail:       adminEmail,
		AdminPass:        adminPass,
	}, nil
}

func (c *Config) NewLimiter() *rate.Limiter {
	limit, burst := 50, 30
	return rate.NewLimiter(rate.Limit(limit), burst)
}
