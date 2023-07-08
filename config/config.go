package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/coreos/go-oidc"
	"gopkg.in/yaml.v2"
)

var (
	providerConfigMap map[string]*ProviderConfig
)

func init() {
	// load config.yaml
	b, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var m map[string]Config

	err = yaml.Unmarshal(b, &m)
	if err != nil {
		log.Fatal(err)
	}

	config, ok := m["dev"]
	if !ok {
		log.Fatal("key dev not found.")
	}
	providerConfigMap = config.ProviderConfigMap
}

type Config struct {
	AppConfig         AppConfig                  `yaml:"server"`
	ProviderConfigMap map[string]*ProviderConfig `yaml:"provider"`
}

type AppConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (c *AppConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type ProviderConfig struct {
	ClientId     string `yaml:"clienId"`
	ClientSecret string `yaml:"clientSecret"`
	Issuer       string `yaml:"issuer"`
	RedirectUrl  string `yaml:"redirectUrl"`
}

// OpenID管理用のProvider型を取得する
func GetProvider(ctx context.Context, providerName string) (*oidc.Provider, error) {
	config, err := GetProviderConfig(providerName)
	if err != nil {
		return nil, err
	}
	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// ProviderConfigを取得する
func GetProviderConfig(providerName string) (*ProviderConfig, error) {
	config, ok := providerConfigMap[providerName]
	if !ok {
		return nil, errors.New("error")
	}
	return config, nil
}
