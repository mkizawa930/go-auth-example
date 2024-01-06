package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// トークンから取得したユーザー情報
type Auth struct {
	UserId string
}

type Jwter struct {
	SecretKey      string
	ExpirationTime time.Duration
}

// jwtトークンを生成する
func GenerateToken() (string, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"user_id": uid.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("secret_key"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// OpenIDによる認証を行う
func Authenticate(ctx context.Context, providerName string, code, nonce string) (bool, error) {
	// get oauth2 config
	oauth2Config, err := GetOAuth2Config(ctx, providerName)
	if err != nil {
		return false, err
	}

	// get token verifier
	verifier, err := GetIDTokenVerifier(ctx, providerName)
	if err != nil {
		return false, err
	}

	// exchange auth code
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Fatal("error")
	}

	// idトークンの検証
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

	if err := idToken.Claims(&claims); err != nil {
		return false, err
	}
	slog.Info("claims", "%+v", claims)

	return true, nil
}

// jwtトークンをparseする
func Parse(signedString string) (*Auth, error) {
	token, err := jwt.Parse(signedString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret_key"), nil
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

// middleware
func AuthenticateMiddleware(next http.Handler) http.Handler {
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
		auth, err := Parse(signedString)
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
