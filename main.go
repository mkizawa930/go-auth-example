package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"oidc_example/config"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
)

//go:embed .secret_key
var SECRET_KEY []byte

type contextKey string

const contextKeyForUserId contextKey = "user_id"

func AuthenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// authenticate
		log.Println("AuthenticateMiddleware is called")
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			respondError(w, http.StatusUnauthorized, errors.New("access_token not found"))
			return
		}
		bearerString := strings.Split(authorizationHeader, " ")

		if len(bearerString) != 2 {
			log.Println(bearerString)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		signedString := bearerString[1]
		auth, err := Parse(signedString)
		log.Printf("%v", signedString)
		if err != nil {
			respondError(w, http.StatusUnauthorized, err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyForUserId, auth.UserId)
		r = r.WithContext(ctx)
		fmt.Println("authorized!")
		next.ServeHTTP(w, r)
	})
}

func main() {
	config.New()

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/{param}", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"param": chi.URLParam(r, "param")})
		})
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/login/{provider}", LoginHandler)
		r.Get("/login/{provider}/callback", CallbackHandler)
	})

	r.Route("/protected", func(r chi.Router) {
		r.Use(AuthenticateMiddleware)
		r.Get("/hello", HelloHandler)
	})

	err := http.ListenAndServe(":8080", r)
	log.Fatal(err)
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, welcome!"})
	w.WriteHeader(http.StatusOK)
}

// ログインハンドラ
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

	state, err := randomString(16)
	if err != nil {
		log.Fatal(err)
	}
	nonce, err := randomString(16)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("state=%v, nonce=%v", state, nonce)
	setCallbackCookie(w, r, "state", state)
	setCallbackCookie(w, r, "nonce", nonce)

	http.Redirect(w, r, oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

// コールバックハンドラ
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")

	fmt.Println(r.Cookies())

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

	ok, err := Authenticate(ctx, "google", code, nonce.Value)
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

// OpenIDによる認証を行う
func Authenticate(ctx context.Context, providerName string, code, nonce string) (bool, error) {
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
	log.Println(rawIdToken)

	idToken, err := verifier.Verify(ctx, rawIdToken)
	if err != nil {
		log.Fatal(err)
	}
	// nonceのチェック
	if idToken.Nonce != nonce {
		return false, errors.New("nonceが一致しません")
	}

	var claims struct {
		Email string `json:"email"`
	}

	if err = idToken.Claims(&claims); err != nil {
		log.Fatal(err)
	}

	return true, nil
}

// jwtトークンを生成する
func GenerateToken() (string, error) {
	userId := "12345"
	claims := jwt.MapClaims{
		"user_id": userId,
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

// トークンから取得したユーザー情報
type Auth struct {
	UserId string
}

// jwtトークンをparseする
func Parse(signedString string) (*Auth, error) {
	token, err := jwt.Parse(signedString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return SECRET_KEY, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, fmt.Errorf("%s is expired", signedString)
			} else {
				return nil, fmt.Errorf("%s is invalid", signedString)
			}
		} else {
			return nil, err
		}
	}

	if token == nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("error when parsing token claims")
	}

	userId, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}

	return &Auth{UserId: userId}, nil
}

// 長さnのランダム文字列(URLエンコード)を生成する
func randomString(n int) (string, error) {
	var b = make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func respondJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(status)
}

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

// func generateToken() {

// }
