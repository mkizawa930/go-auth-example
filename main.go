package main

import (
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type loggerCtxKey struct{}

func init() {
	// TODO:
}

func main() {
	c := new(Config) // TODO

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	slog.Debug("start main")
	LoadConfig()

	loggerMiddleware := NewLoggerMiddleware(logger)
	authMiddleware := NewAuthMiddleware(c)

	callbackHandler := NewCallbackHandlerFunc(c)
	authHandler := NewAuthHandler(c)

	router := NewRouter(loggerMiddleware, authMiddleware, authHandler, callbackHandler)

	addr := c.Addr()
	slog.Info(fmt.Sprintf("start listening on %s", addr))

	// start servers
	log.Fatal(http.ListenAndServe(c.Addr(), router))
}
