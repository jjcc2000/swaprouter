package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/jjcc2000/swaprouter/internal/adapters/oneinch"
	"github.com/jjcc2000/swaprouter/internal/aggregator"
	"github.com/jjcc2000/swaprouter/internal/config"
	"github.com/jjcc2000/swaprouter/internal/db"
	"github.com/jjcc2000/swaprouter/internal/gateway/handlers"
	"github.com/jjcc2000/swaprouter/internal/gateway/middleware"
	"github.com/jjcc2000/swaprouter/internal/repository"
	"github.com/jjcc2000/swaprouter/pkg/logger"
)

func main() {
	godotenv.Load()
	cfg, err := config.Load()
	
	if err != nil {
		log.Fatal("config error:", err)
	}

	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("db error:", err)
	}

	tradeRepo := repository.NewTradeRepository(database)

	log := logger.New(cfg.Env)

	// Redis
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Error("redis url invalid", "err", err)
	}
	rdb := redis.NewClient(opt)

	// Adapters — empty for now, added in next layer
	adapters := []aggregator.IAdapter{
		oneinch.New(cfg.OneInchAPIKey,cfg.OneInchBaseURL),
	}

	engine := aggregator.NewQuoteEngine(adapters, cfg.QuoteTimeoutMs)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	auth := handlers.NewAuthHandler(rdb, cfg.JWTSecret)

	r.Get("/auth/nonce", auth.Nonce)
	r.Post("/auth/login", auth.Login)

	// Public routes
	r.Get("/health", handlers.HealthHandler)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Use(middleware.RateLimiter(rdb, cfg.RateLimitRPM))

		r.Get("/v1/quote", handlers.NewQuoteHandler(engine).ServeHTTP)
		r.Post("/v1/swap", handlers.NewSwapHandler(engine, tradeRepo).ServeHTTP)
		r.Get("/v1/trades", handlers.NewTradesHandler(tradeRepo).ServeHTTP)
		r.Get("/v1/tokens", handlers.TokensHandler)
		r.Get("/v1/chains", handlers.ChainsHandler)
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info("SwapRouter running", "addr", addr)
	http.ListenAndServe(addr, r)
}
