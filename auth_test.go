package main

import (
	"testing"
)

func Test_GenerateToken(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	_, err = Parse(token)
	if err != nil {
		t.Fatal(err)
	}
}
