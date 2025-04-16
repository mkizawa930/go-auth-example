package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
)

// 認可コードを使った認証フローの実装
// 1. 認可コードを受取り、IDプロバイダへ認可コードを送信し、IDトークンを受け取る
// 2. 受け取ったIDトークンを検証する
func Authenticate(ctx context.Context, c *Config, providerName string, code string, nonce string) (map[string]any, error) {

	// get oauth2 config
	oauth2Config, err := NewOAuth2Config(ctx, c, providerName)
	if err != nil {
		return nil, err
	}

	// exchange auth code
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("failed to get id_token")
	}

	// get token verifier
	verifier, err := NewIdTokenVerifier(ctx, c, providerName)
	if err != nil {
		return nil, err
	}

	// idトークンの検証
	idToken, err := verifier.Verify(ctx, rawIdToken)
	if err != nil {
		return nil, err
	}

	// nonceのチェック
	if idToken.Nonce != nonce {
		return nil, errors.New("nonceが一致しません")
	}

	var claims = make(map[string]any)
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// jwtトークンを生成する
func GenerateToken(c *Config, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub": email,
		"iss": "issuer", // TODO
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(c.JwtSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// jwtトークンをparseする
func ParseToken(c *Config, signedString string) (jwt.MapClaims, error) {
	// トークンの暗号を検証する
	token, err := jwt.Parse(signedString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.JwtSecretKey), nil
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

	return claims, nil
}
