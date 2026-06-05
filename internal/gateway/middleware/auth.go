package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const WalletKey contextKey = "wallet"

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				writeUnauthorized(w, "missing authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				writeUnauthorized(w, "invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeUnauthorized(w, "invalid claims")
				return
			}

			wallet, ok := claims["wallet"].(string)
			if !ok || wallet == "" {
				writeUnauthorized(w, "missing wallet claim")
				return
			}

			ctx := context.WithValue(r.Context(), WalletKey, wallet)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WalletFromContext(ctx context.Context) string {
	wallet, _ := ctx.Value(WalletKey).(string)
	return wallet
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"code":"UNAUTHORIZED","message":"` + msg + `"}`))
}