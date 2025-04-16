package main

import (
	"crypto/rand"
	"encoding/base64"
)

// 長さnのランダム文字列(URLエンコード)を生成する
func randomString(n int) (string, error) {
	var b = make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
