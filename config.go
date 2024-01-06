package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/coreos/go-oidc"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	cfg               *ServerConfig
	providerConfigMap map[string]*ProviderConfig
	oauth2ConfigMap   map[string]*oauth2.Config
	verifierMap       map[string]*oidc.IDTokenVerifier
)

type ProviderName string

const (
	Google string = "google"
)

type ServerConfig struct {
	Host string `env:"SERVER_HOST"`
	Port int    `env:"SERVER_PORT"`
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	Issuer       string
	RedirectUrl  string
	Scopes       []string
}

func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func LoadConfig() {
	godotenv.Load()
	// init ServerConfig
	cfg = &ServerConfig{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	// init providerConfigMap
	providerConfigMap = make(map[string]*ProviderConfig)

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

// type ProviderConfig struct {
// 	ClientId     string
// 	ClientSecret string
// 	Issuer       string
// 	RedirectUrl  string
// }

// ProviderConfigを取得する
// func GetProviderConfig(providerName string) (*ProviderConfig, error) {
// 	config, ok := providerConfigMap[providerName]
// 	if !ok {
// 		return nil, fmt.Errorf("%vというプロバイダ情報が見つかりません", providerName)
// 	}
// 	return config, nil
// }

// func GetOAuth2Config(ctx context.Context, providerName string) (*oauth2.Config, error) {
// 	var err error
// 	cfg, ok := providerConfigMap[providerName]
// 	if !ok {
// 		return nil, fmt.Errorf("provider config not found")
// 	}

// 	provider, err := oidc.NewProvider(ctx, cfg.GetIssuer())
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &oauth2.Config{
// 		ClientID:     cfg.ClientId,
// 		ClientSecret: cfg.ClientSecret,
// 		Endpoint:     provider.Endpoint(),
// 		RedirectURL:  cfg.RedirectUrl,
// 		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
// 	}, nil
// }

func GetOAuth2Config(ctx context.Context, providerName string) (*oauth2.Config, error) {

	cfg, ok := providerConfigMap[providerName]
	if !ok {
		return nil, fmt.Errorf("%vの設定情報が見つかりません", providerName)
	}

	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		slog.Error(err.Error())
		return nil, fmt.Errorf("oidc.Providerの作成に失敗しました")
	}

	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectUrl,
		Scopes:       cfg.Scopes,
	}, nil
}

// func SetVerifier(provider string, verifier *oidc.IDTokenVerifier) bool {
// 	if _, ok := verifierMap[provider]; !ok {
// 		verifierMap[provider] = verifier
// 		return true
// 	}
// 	return false
// }

func GetIDTokenVerifier(ctx context.Context, providerName string) (*oidc.IDTokenVerifier, error) {
	cfg, ok := providerConfigMap[providerName]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("%sに関するプロバイダ情報が見つかりません", providerName))
	}
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		slog.Error(err.Error())
		return nil, fmt.Errorf("oidc.Providerの生成に失敗しました")
	}

	return provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}), nil
}
