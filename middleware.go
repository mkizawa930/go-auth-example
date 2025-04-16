package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
)

// ロガーを注入するミドルウェア
func NewLoggerMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("LoggerMiddleware is called")
			next.ServeHTTP(w, r)
		})
	}
}

// 認証ミドルウェア
func NewAuthMiddleware(c *Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// authenticate
			log.Println("AuthenticateMiddleware is called")
			authorizationHeader := r.Header.Get("Authorization")
			if authorizationHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				respondError(w, http.StatusUnauthorized, errors.New("access token is not found"))
				return
			}
			bearerString := strings.Split(authorizationHeader, " ")

			if len(bearerString) != 2 {
				log.Println(bearerString)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			signedString := bearerString[1]
			claims, err := ParseToken(c, signedString)
			if err != nil {
				respondError(w, http.StatusUnauthorized, err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID, ok := claims["sub"].(string)
			if !ok {
				// TODO
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// contextにUserIDをセットする
			ctx := context.WithValue(r.Context(), &userIdCtxKey{}, userID)
			r = r.WithContext(ctx)

			fmt.Println("authorized!") // TODO: loggerに置き換える

			next.ServeHTTP(w, r)
		})
	}
}
