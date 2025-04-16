package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/coreos/go-oidc"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type Provider string

const (
	ProviderGoogle string = "google"
)

type Config struct {
	ServerHost        string `env:"SERVER_HOST"`
	ServerPort        string `env:"SERVER_PORT"`
	JwtSecretKey      string `env:"JWT_SECRET_KEY"`
	ProviderConfigMap map[string]*ProviderConfig
	ProviderMap       map[string]*oidc.Provider
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.ServerHost, c.ServerPort)
}

type ServerConfig struct {
	Host string `env:"SERVER_HOST"`
	Port int    `env:"SERVER_PORT"`
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	Issuer       string
	RedirectUrl  string
	Endpoint     oauth2.Endpoint
	Scopes       []string
}

func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func LoadConfig() {
	godotenv.Load()
	// init ServerConfig
	c := new(Config)
	if err := env.Parse(c); err != nil {
		panic(err)
	}

	// init providerConfigMap
	providerConfigMap := make(map[string]*ProviderConfig)

	providers := []string{"google"}
	for _, provider := range providers {
		name := strings.ToUpper(provider)
		clientIdKey := fmt.Sprintf("%s_CLIENT_ID", name)
		clientSecretKey := fmt.Sprintf("%s_CLIENT_SECRET", name)
		issuerKey := fmt.Sprintf("%s_ISSUER", name)
		redirectUrlKey := fmt.Sprintf("%s_REDIRECT_URL", name)
		scopesKey := fmt.Sprintf("%s_SCOPES", name)

		clientID := os.Getenv(clientIdKey)
		if len(clientID) == 0 {
			panic(fmt.Sprintf("%s is missing", clientIdKey))
		}
		clientSecret := os.Getenv(clientSecretKey)
		if len(clientSecret) == 0 {
			panic(fmt.Sprintf("%s is missing", clientSecretKey))
		}
		issuer := os.Getenv(issuerKey)
		if len(issuer) == 0 {
			panic(fmt.Sprintf("%s is missing", issuerKey))
		}
		redirectUrl := os.Getenv(redirectUrlKey)
		if len(redirectUrl) == 0 {
			panic(fmt.Sprintf("%s is missing", redirectUrlKey))
		}
		scopesRawString := os.Getenv(scopesKey)
		scopes := strings.Split(scopesRawString, ",")

		providerConfigMap[provider] = &ProviderConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Issuer:       issuer,
			RedirectUrl:  redirectUrl,
			Scopes:       scopes,
		}
	}
}

// OAuth2Configを取得する
func (c *Config) NewOAuth2Config(ctx context.Context, key string) (*oauth2.Config, error) {
	providerConfig, ok := c.ProviderConfigMap[key]
	if !ok {
		return nil, fmt.Errorf("プロバイダ設定が見つかりません: key=%s", key)
	}

	// provider, err := oidc.NewProvider(context.Background(), providerConfig.Issuer)
	// if err != nil {
	// 	slog.Error(err.Error())
	// 	return nil, fmt.Errorf("oidc.Providerの作成に失敗しました")
	// }

	return &oauth2.Config{
		ClientID:     providerConfig.ClientID,
		ClientSecret: providerConfig.ClientSecret,
		Endpoint:     providerConfig.Endpoint, // TODO
		RedirectURL:  providerConfig.RedirectUrl,
		Scopes:       providerConfig.Scopes,
	}, nil
}

func (c *Config) GetProvider(ctx context.Context, key string) (*oidc.Provider, error) {
	provider, ok := c.ProviderMap[key]
	if ok {
		return provider, nil
	}

	providerConfig, ok := c.ProviderConfigMap[key]
	if !ok {
		return nil, fmt.Errorf("プロバイダ設定が見つかりません: key=%s", key)
	}

	provider, err := oidc.NewProvider(ctx, providerConfig.Issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc.Providerの作成に失敗しました")
	}
	c.ProviderMap[key] = provider

	return provider, nil
}

// TokenVerifierを取得する
func (c *Config) NewIdTokenVerifier(ctx context.Context, key string) (*oidc.IDTokenVerifier, error) {
	providerConfig, ok := c.ProviderConfigMap[key]
	if !ok {
		return nil, fmt.Errorf("プロバイダ設定が見つかりません: key=%s", key)
	}
	provider, err := c.GetProvider(ctx, key)
	if err != nil {
		return nil, err
	}
	return provider.Verifier(&oidc.Config{ClientID: providerConfig.ClientID}), nil
}

// func SetVerifier(provider string, verifier *oidc.IDTokenVerifier) bool {
// 	if _, ok := verifierMap[provider]; !ok {
// 		verifierMap[provider] = verifier
// 		return true
// 	}
// 	return false
// }
