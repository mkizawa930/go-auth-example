package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

type userIdCtxKey struct{}

type Middleware func(http.Handler) http.Handler

func NewRouter(
	loggerMiddleware,
	authMiddleware Middleware,
	authHandler http.HandlerFunc,
	callbackHandler http.HandlerFunc,
) chi.Router {
	r := chi.NewRouter()

	r.Use(loggerMiddleware)

	r.Route("/test", func(r chi.Router) {
		r.Get("/", indexHandler)
	})

	r.Route("/auth/{provider}", func(r chi.Router) {
		r.Get("/", authHandler)
		r.Get("/callback", callbackHandler)
	})

	r.Route("/protected", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/", protectedHandler)
	})
	return r
}
