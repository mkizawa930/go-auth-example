# 外部プロバイダを使ったトークン認証のGo実装サンプル

## 使用したライブラリ

github.com/go-chi/chi
軽量なWebフレームワーク

golang.org/x/oauth2 
OAuth2を扱う

github.com/golang-jwt/jwt
JWTトークンを扱う

github.com/coreos/go-oidc
IDトークンの検証などに使用


## 起動

`/auth/{provider}`: 認可コード発行URLの取得
`/auth/{provider}/callback`: 認可コードを受け取るエンドポイント


```
go build . && ./oidc_example

```



