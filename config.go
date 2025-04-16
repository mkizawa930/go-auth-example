package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/coreos/go-oidc"
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
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.ServerHost, c.ServerPort)
}

// providerをキャッシュすると内部エラーが起きるので、都度生成する
func (c *Config) GetProvider(ctx context.Context, key string) (*oidc.Provider, error) {
	providerConfig, ok := c.ProviderConfigMap[key]
	if !ok {
		return nil, fmt.Errorf("プロバイダ設定が見つかりません: key=%s", key)
	}

	provider, err := oidc.NewProvider(ctx, providerConfig.Issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc.Providerの作成に失敗しました")
	}

	return provider, nil
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	Issuer       string
	RedirectUrl  string
	Scopes       []string
}

func InitConfig() (*Config, error) {
	// init ServerConfig
	c := &Config{}

	if err := env.Parse(c); err != nil {
		panic(err)
	}

	providers := []string{
		"google",
	}
	providerConfigMap, err := initProviderConfig(providers)
	if err != nil {
		return nil, err
	}

	c.ProviderConfigMap = providerConfigMap

	return c, nil
}

func initProviderConfig(providers []string) (map[string]*ProviderConfig, error) {
	// init providerConfigMap
	providerConfigMap := make(map[string]*ProviderConfig)

	for _, provider := range providers {

		prefix := strings.ToUpper(provider)

		clientIDKey := fmt.Sprintf("%s_CLIENT_ID", prefix)
		clientID := os.Getenv(clientIDKey)
		if len(clientID) == 0 {
			return nil, fmt.Errorf("%sが見つかりません", clientIDKey)
		}

		clientSecretKey := fmt.Sprintf("%s_CLIENT_SECRET", prefix)
		clientSecret := os.Getenv(clientSecretKey)
		if len(clientSecret) == 0 {
			return nil, fmt.Errorf("%sが見つかりません", clientSecretKey)
		}

		issuerKey := fmt.Sprintf("%s_ISSUER", prefix)
		issuer := os.Getenv(fmt.Sprintf("%s_ISSUER", prefix))
		if len(issuer) == 0 {
			panic(fmt.Sprintf("%sが見つかりません", issuerKey))
		}

		redirectUrlKey := fmt.Sprintf("%s_REDIRECT_URL", prefix)
		redirectUrl := os.Getenv(redirectUrlKey)
		if len(redirectUrl) == 0 {
			panic(fmt.Sprintf("%sが見つかりません", redirectUrlKey))
		}

		scopesKey := fmt.Sprintf("%s_SCOPES", prefix)
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

	return providerConfigMap, nil
}

// OAuth2Configを取得する
func NewOAuth2Config(ctx context.Context, c *Config, key string) (*oauth2.Config, error) {
	provider, err := c.GetProvider(ctx, key)
	if err != nil {
		return nil, err
	}

	providerConfig, ok := c.ProviderConfigMap[key]
	if !ok {
		return nil, fmt.Errorf("プロバイダ設定が見つかりません: key=%s", key)
	}

	return &oauth2.Config{
		ClientID:     providerConfig.ClientID,
		ClientSecret: providerConfig.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  providerConfig.RedirectUrl,
		Scopes:       providerConfig.Scopes,
	}, nil
}

// TokenVerifierを取得する
func NewIdTokenVerifier(ctx context.Context, c *Config, key string) (*oidc.IDTokenVerifier, error) {
	provider, err := c.GetProvider(ctx, key)
	if err != nil {
		return nil, err
	}

	providerConfig, ok := c.ProviderConfigMap[key]
	if !ok {
		return nil, fmt.Errorf("プロバイダ設定が見つかりません: key=%s", key)
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
