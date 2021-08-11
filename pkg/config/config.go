package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Debug                 bool
	Addr                  string
	RequestBodyLimitBytes int64
	RequestTimeout        time.Duration
	IdleTimeout           time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	PostTTL               time.Duration
	CORSAllowOrigins      []string
	DatabaseName          string
	DatabaseURL           string
	JWTSecret             string
}

func NewFromEnv(filenames ...string) (*Config, error) {

	if err := godotenv.Load(filenames...); err != nil {
		return nil, fmt.Errorf("failed to get env variables: %w", err)
	}

	host := os.Getenv("HOST")
	if host == "" {
		return nil, errors.New("no HOST env variable set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, errors.New("no PORT env variable set")
	}

	reqBodyLimit, err := strconv.ParseInt(os.Getenv("REQBODYLIMIT_BYTES"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse REQBODYLIMIT_BYTES env variable to int64: %w", err)
	}

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DEBUG env variable to bool: %w", err)
	}

	reqTimeout, err := time.ParseDuration(os.Getenv("TIMEOUT_REQ"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse TIMEOUT_REQ env variable to time.Duration: %w", err)
	}

	idleTimeout, err := time.ParseDuration(os.Getenv("TIMEOUT_IDLE"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse TIMEOUT_IDLE env variable to time.Duration: %w", err)
	}

	readTimeout, err := time.ParseDuration(os.Getenv("TIMEOUT_READ"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse TIMEOUT_READ env variable to time.Duration: %w", err)
	}

	writeTimeout, err := time.ParseDuration(os.Getenv("TIMEOUT_WRITE"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse TIMEOUT_WRITE env variable to time.Duration: %w", err)
	}

	postTTL, err := time.ParseDuration(os.Getenv("POST_TTL"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse POST_TTL env variable to time.Duration: %w", err)
	}

	corsAllowOrigins := strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")

	dbName := os.Getenv("DATABASE_NAME")
	if dbName == "" {
		return nil, errors.New("no DATABASE_NAME env variable set")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("no DATABASE_URL env variable set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("no JWT_SECRET env variable set")
	}

	return &Config{
		Debug:                 debug,
		Addr:                  fmt.Sprintf("%s:%s", host, port),
		RequestBodyLimitBytes: reqBodyLimit,
		RequestTimeout:        reqTimeout,
		IdleTimeout:           idleTimeout,
		ReadTimeout:           readTimeout,
		WriteTimeout:          writeTimeout,
		CORSAllowOrigins:      corsAllowOrigins,
		PostTTL:               postTTL,
		DatabaseName:          dbName,
		DatabaseURL:           dbURL,
		JWTSecret:             jwtSecret,
	}, nil
}
