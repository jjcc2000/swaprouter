package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port            string
	Env             string
	JWTSecret       string
	RedisURL        string
	DatabaseURL     string
	QuoteTTLSeconds int
	RateLimitRPM    int
	QuoteTimeoutMs  int
	UniswapRPCURL   string
	OneInchAPIKey   string
	OneInchBaseURL  string
	JupiterBaseURL  string
	SolanaRPCURL    string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "development"),
		JWTSecret:      mustEnv("JWT_SECRET"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		DatabaseURL:    mustEnv("DATABASE_URL"),
		UniswapRPCURL:  getEnv("UNISWAP_RPC_URL", ""),
		OneInchAPIKey:  getEnv("ONEINCH_API_KEY", ""),
		OneInchBaseURL: getEnv("ONEINCH_BASE_URL", "https://api.1inch.dev/swap/v6.0"),
		JupiterBaseURL: getEnv("JUPITER_BASE_URL", "https://quote-api.jup.ag/v6"),
		SolanaRPCURL:   getEnv("SOLANA_RPC_URL", "https://api.mainnet-beta.solana.com"),
	}

	var err error
	cfg.QuoteTTLSeconds, err = getEnvInt("QUOTE_TTL_SECONDS", 5)
	if err != nil {
		return nil, fmt.Errorf("QUOTE_TTL_SECONDS: %w", err)
	}
	cfg.RateLimitRPM, err = getEnvInt("RATE_LIMIT_RPM", 60)
	if err != nil {
		return nil, fmt.Errorf("RATE_LIMIT_RPM: %w", err)
	}
	cfg.QuoteTimeoutMs, err = getEnvInt("QUOTE_TIMEOUT_MS", 300)
	if err != nil {
		return nil, fmt.Errorf("QUOTE_TIMEOUT_MS: %w", err)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}

func getEnvInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	return strconv.Atoi(v)
}