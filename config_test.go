package main

import (
	"context"
	"testing"

	"github.com/coreos/go-oidc"
)

func Test_Provider(t *testing.T) {
	issuer := "https://accounts.google.com"
	provider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		t.Fatal(err)
	}
	var claims = make(map[string]any)
	if err := provider.Claims(&claims); err != nil {
		t.Fatal(err)
	}
	t.Log(provider.Endpoint())
	// b, _ := json.MarshalIndent(claims, "", "  ")

	// t.Log(string(b))
}
