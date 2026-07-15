package handler

import (
	"context"
	"net"
	"net/http"
	"time"

	"IM_Chat_System/internal/auth"
	"IM_Chat_System/internal/httpx"
	"IM_Chat_System/internal/ratelimit"
	"IM_Chat_System/internal/tokenblacklist"
)

type contextKey string

const claimsKey contextKey = "claims"
const tokenKey contextKey = "token"

func WithAuth(secret string, blacklist tokenblacklist.Store, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.BearerToken(r.Header.Get("Authorization"))
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "missing token")
			return
		}
		if blacklist != nil {
			blocked, err := blacklist.Contains(r.Context(), token)
			if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "check token blacklist failed")
				return
			}
			if blocked {
				httpx.WriteError(w, http.StatusUnauthorized, "token has been revoked")
				return
			}
		}

		claims, err := auth.ParseToken(secret, token)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		ctx = context.WithValue(ctx, tokenKey, token)
		next(w, r.WithContext(ctx))
	}
}

func ClaimsFromContext(ctx context.Context) (auth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(auth.Claims)
	return claims, ok
}

func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey).(string)
	return token, ok
}

func WithRateLimit(store ratelimit.Store, scope string, limit int64, window time.Duration, keyFn func(*http.Request) string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			next(w, r)
			return
		}
		key := scope + ":" + keyFn(r)
		allowed, err := store.Allow(r.Context(), key, limit, window)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "rate limit check failed")
			return
		}
		if !allowed {
			httpx.WriteError(w, http.StatusTooManyRequests, "too many requests")
			return
		}
		next(w, r)
	}
}

func ClientIPKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
