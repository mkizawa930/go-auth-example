package main

import (
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type contextKey string

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	slog.Debug("start main")
	LoadConfig()

	r := NewRouter()
	addr := cfg.Addr()
	slog.Info(fmt.Sprintf("start listening on %s", addr))
	err := http.ListenAndServe(addr, r)
	log.Fatal(err)
}
