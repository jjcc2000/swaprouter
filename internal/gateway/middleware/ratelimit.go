package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

func RateLimiter(rdb *redis.Client, rpm int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wallet := WalletFromContext(r.Context())
			if wallet == "" {
				next.ServeHTTP(w, r)
				return
			}

			key := fmt.Sprintf("rate:%s:%d", wallet, time.Now().Minute())
			ctx := context.Background()

			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				rdb.Expire(ctx, key, 2*time.Minute)
			}

			if int(count) > rpm {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"code":"RATE_LIMITED","message":"too many requests"}`))
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rpm))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rpm-int(count)))
			next.ServeHTTP(w, r)
		})
	}
}