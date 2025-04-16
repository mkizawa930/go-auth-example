package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/go-chi/chi"
)

// ルートハンドラ
func indexHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, world!"})
	w.WriteHeader(http.StatusOK)
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"message": "success"})
}

// 認証ハンドラ
// 認可コード発行先のURLを返す
func NewAuthHandler(c *Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		providerName := chi.URLParam(r, "provider")
		if providerName == "" {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("provider is not found"))
			return
		}

		oauth2Config, err := NewOAuth2Config(ctx, c, providerName)
		if err != nil {
			slog.Error(err.Error())
			respondError(w, http.StatusInternalServerError, nil)
			return
		}

		// stateの設定
		state, err := randomString(16)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}
		setCallbackCookie(w, r, "state", state)

		// nonceの設定
		nonce, err := randomString(16)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}
		setCallbackCookie(w, r, "nonce", nonce)

		authCodeURL := oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce))

		http.Redirect(w, r, authCodeURL, http.StatusFound)
	})
}

// コールバックハンドラ
func NewAuthCallbackHandlerFunc(c *Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger, ok := ctx.Value(&loggerCtxKey{}).(*slog.Logger)
		if !ok {
			logger = slog.Default() // TODO:
		}

		providerName := chi.URLParam(r, "provider")
		code := r.URL.Query().Get("code") // 認可コード

		state, err := r.Cookie("state")
		if err != nil {
			http.Error(w, "state not found", http.StatusBadRequest)
			return
		}

		nonce, err := r.Cookie("nonce")
		if err != nil {
			http.Error(w, "nonce not found", http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("state") != state.Value {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		// 認可コードフロー
		claims, err := Authenticate(ctx, c, providerName, code, nonce.Value)
		if err != nil {
			log.Println(err)
			return
		}

		email, ok := claims["email"].(string)
		if !ok {
			logger.Error("invalid claims")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// アクセストークンの生成
		accessToken, err := GenerateToken(c, email)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("アクセストークンの生成に失敗しました"))
			return
		}

		// accessTokenを返す
		respBody := map[string]any{
			"accessToken": accessToken,
			"tokenType":   "bearer",
		}
		respondJSON(w, http.StatusOK, respBody)
	})
}
