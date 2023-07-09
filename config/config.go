package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

var (
	appConfig         *AppConfig
	providerConfigMap map[string]*ProviderConfig
	verifierMap       map[string]*oidc.IDTokenVerifier
)

func init() {
	verifierMap = make(map[string]*oidc.IDTokenVerifier)
}

func New() {
	// load config.yaml
	m := loadYaml()
	config, ok := m["dev"]
	if !ok {
		log.Fatal("key dev not found.")
	}
	// プロバイダ情報をコピー
	appConfig = config.AppConfig
	providerConfigMap = config.ProviderConfigMap
}

// yamlの設定ファイルを読みこむ
func loadYaml() map[string]Config {
	dir, _ := os.Getwd()
	fmt.Println(dir)
	b, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var m map[string]Config

	err = yaml.Unmarshal(b, &m)
	if err != nil {
		log.Fatal(err)
	}

	return m
}

// 初期セットアップ時実行関数
func Setup() {

}

type Config struct {
	AppConfig         *AppConfig                 `yaml:"server"`
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
	ClientId     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
	Issuer       string `yaml:"issuer"`
	RedirectUrl  string `yaml:"redirectUrl"`
}

// ProviderConfigを取得する
func GetProviderConfig(providerName string) (*ProviderConfig, error) {
	config, ok := providerConfigMap[providerName]
	if !ok {
		return nil, fmt.Errorf("%vというプロバイダ情報が見つかりません", providerName)
	}
	return config, nil
}

func GetOAuth2Config(ctx context.Context, providerName string, config *ProviderConfig) (*oauth2.Config, error) {
	var err error
	if config == nil {
		config, err = GetProviderConfig(providerName)
		if err != nil {
			return nil, err
		}
	}

	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		return nil, err
	}
	fmt.Println(config.ClientId)
	fmt.Println(config.ClientSecret)
	return &oauth2.Config{
		ClientID:     config.ClientId,
		ClientSecret: config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  config.RedirectUrl,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}, nil
}

func SetVerifier(provider string, verifier *oidc.IDTokenVerifier) bool {
	if _, ok := verifierMap[provider]; !ok {
		verifierMap[provider] = verifier
		return true
	}
	return false
}

func GetVerifier(provider string) (*oidc.IDTokenVerifier, error) {
	verifier, ok := verifierMap[provider]
	if !ok {
		return nil, fmt.Errorf("verifier for %v is not found", provider)
	}
	return verifier, nil
}
