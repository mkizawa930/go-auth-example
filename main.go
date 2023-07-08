package main

import (
	"context"
	"fmt"
	"net/http"
	"oidc_example/config"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func init() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", IndexHandler)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	v, err := config.GetProvider(context.Background(), "google")
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}
