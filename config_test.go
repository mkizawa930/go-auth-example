package main

import (
	"context"
	"testing"

	"github.com/coreos/go-oidc"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
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

func Test_InitConfig(t *testing.T) {

	t.Run("環境変数から設定の読み取りに成功する", func(t *testing.T) {

		if err := godotenv.Load(".env"); err != nil {
			t.Fatal(err)
		}

		c, err := InitConfig()
		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, c.ServerHost)
		assert.NotEmpty(t, c.ServerPort)

		for _, v := range c.ProviderConfigMap {
			assert.NotEmpty(t, v)
		}
	})

}
