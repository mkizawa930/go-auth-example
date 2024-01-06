package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/go-chi/chi"
)

const contextKeyForUserId contextKey = "user_id"

func NewRouter() chi.Router {
	r := chi.NewRouter()

	// r.Route("/", func(r chi.Router) {
	// 	r.Get("/{param}", func(w http.ResponseWriter, r *http.Request) {
	// 		json.NewEncoder(w).Encode(map[string]string{"param": chi.URLParam(r, "param")})
	// 	})
	// })
	r.Route("/", func(r chi.Router) {
		r.Get("/hello", helloHandler)
	})

	r.Route("/auth/{provider}", func(r chi.Router) {
		r.Get("/", authHandler)
		r.Get("/callback", callbackHandler)
	})

	r.Route("/protected", func(r chi.Router) {
		r.Use(AuthenticateMiddleware)
		r.Get("/hello", helloHandler)
	})
	return r
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("helloHandler is called")
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, welcome!"})
	w.WriteHeader(http.StatusOK)
}

// 認証ハンドラ
func authHandler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("authHanlder is called")
	providerName := chi.URLParam(r, "provider")
	if providerName == "" {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("provider is not found"))
		return
	}
	ctx := r.Context()
	cfg, err := GetOAuth2Config(ctx, providerName)
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

	http.Redirect(w, r, cfg.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

// コールバックハンドラ
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providerName := chi.URLParam(r, "provider")
	code := r.URL.Query().Get("code") // 認可コード
	slog.Debug("authCode", "code", code)
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
	ok, err := Authenticate(ctx, providerName, code, nonce.Value)
	if err != nil {
		log.Println(err)
		return
	}
	if !ok {
		respondError(w, http.StatusBadRequest, fmt.Errorf("認証に失敗しました"))
		return
	}

	// アクセストークンの生成
	accessToken, err := GenerateToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("アクセストークンの生成に失敗しました"))
		return
	}

	// accessTokenを返す
	resp := struct {
		AccessToken string `json:"accessToken"`
	}{
		AccessToken: accessToken,
	}
	respondJSON(w, http.StatusOK, resp)
}

// 長さnのランダム文字列(URLエンコード)を生成する
func randomString(n int) (string, error) {
	var b = make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// JSONレスポンスを返す
func respondJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(status)
}

// cookieに値をセットする
func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}
