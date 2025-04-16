package main

import (
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type loggerCtxKey struct{}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	slogHandlerOpts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, slogHandlerOpts))
	slog.SetDefault(logger)

	//
	c, err := InitConfig()
	if err != nil {
		logger.Error(err.Error())
		log.Fatal(err)
	}

	loggerMiddleware := NewLoggerMiddleware(logger)
	authMiddleware := NewAuthMiddleware(c)

	authHandler := NewAuthHandler(c)
	authCallbackHandler := NewAuthCallbackHandlerFunc(c)

	router := NewRouter(loggerMiddleware, authMiddleware, authHandler, authCallbackHandler)

	fmt.Printf("start listening: %s\n", c.Addr())

	// start servers
	log.Fatal(http.ListenAndServe(c.Addr(), router))
}
