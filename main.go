package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"oidc_example/config"
	"os"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
)

//go:embed .secret_key
var SECRET_KEY []byte

func main() {
	config.New()

	r := chi.NewRouter()
	r.Get("/{param}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"param": chi.URLParam(r, "param")})
	})
	r.Get("/auth/login/{provider}", LoginHandler)
	r.Get("/auth/login/{provider}/callback", CallbackHandler)

	http.ListenAndServe(":8080", r)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providerName := chi.URLParam(r, "provider")
	if providerName == "" {
		log.Println("providerName is empty")
		return
	}
	oauth2Config, err := config.GetOAuth2Config(ctx, "google", nil)
	if err != nil {
		log.Println(err)
		log.Println("oauth2Configの取得に失敗しました")
	}

	state, err := randomString(32)
	if err != nil {
		log.Fatal(err)
	}
	url := oauth2Config.AuthCodeURL(state)
	err = json.NewEncoder(w).Encode(map[string]string{"redirectUrl": url})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")
	ok, err := Authenticate(ctx, "google", code)
	if err != nil {
		log.Println(err)
		return
	}
	if !ok {
		panic("error")
	}
	accessToken, err := GenerateToken()
	if err != nil {
		return
	}
	resp := struct {
		AccessToken string `json:"accessToken"`
	}{
		AccessToken: accessToken,
	}
	respondJSON(w, http.StatusOK, resp)
}

func Authenticate(ctx context.Context, providerName string, code string) (bool, error) {
	providerConfig, err := config.GetProviderConfig(providerName)
	if err != nil {
		return false, err
	}

	oauth2Config, err := config.GetOAuth2Config(ctx, providerName, providerConfig)
	if err != nil {
		log.Fatal(err)
	}

	provider, err := oidc.NewProvider(ctx, providerConfig.Issuer)
	if err != nil {
		log.Fatal(err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: oauth2Config.ClientID})

	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Fatal("error")
	}

	idToken, err := verifier.Verify(ctx, rawIdToken)
	if err != nil {
		log.Fatal(err)
	}

	var claims struct {
		Email string `json:"email"`
	}

	if err = idToken.Claims(&claims); err != nil {
		log.Fatal(err)
	}

	return true, nil
}

func GenerateToken() (string, error) {
	claims := jwt.MapClaims{
		"user_id": 12345,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	json.NewEncoder(os.Stdin).Encode(token.Header)
	json.NewEncoder(os.Stdin).Encode(token.Claims)

	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func randomString(n int) (string, error) {
	var b = make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return string(b), nil
}

func respondJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(status)
}

// func generateToken() {

// }
