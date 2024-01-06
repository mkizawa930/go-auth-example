package main

import (
	"testing"

	"github.com/go-playground/assert"
)

func Test_ServerConfig(t *testing.T) {
	cfg := ServerConfig{Host: "localhost", Port: 8080}
	assert.Equal(t, cfg.Addr(), "localhost:8080")
}

func Test_LoadConfig(t *testing.T) {
	LoadConfig()

	if cfg == nil {
		t.Fail()
	}

	googleConfig, ok := providerConfigMap["google"]
	if !ok {
		t.Fail()
	}
	t.Log(googleConfig.Scopes)
	assert.Equal(t, len(googleConfig.Scopes), 3)
}

// func TestGetProviderConfig(t *testing.T) {
// 	dirpath, err := os.Getwd()
// 	if err != nil {
// 		t.Log(dirpath)
// 		t.Fail()
// 	}
// 	filepath := dirpath + "/" + "config.yaml"
// 	t.Log(filepath)
// 	b, err := os.ReadFile(filepath)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	var m map[string]Config
// 	err = yaml.Unmarshal(b, &m)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	got, ok := m["dev"]
// 	if ok {
// 		t.Fail()
// 	}
// 	provider, ok := got.ProviderConfigMap["google"]
// 	if !ok {
// 		t.Fail()
// 	}
// 	if provider.Issuer != "https://accounts.google.com" {
// 		t.Log(provider.Issuer)
// 		t.Fail()
// 	}
// }
