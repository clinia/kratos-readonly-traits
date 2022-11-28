package main

import (
	"context"
	"errors"
	"github.com/lus/kratos-readonly-traits/internal/config"
	"github.com/lus/kratos-readonly-traits/internal/webhook"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Set up the logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if cfg.IsDevEnv() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Warn().Msg("The service was started in development mode. Please change the 'ENVIRONMENT' variable to 'prod' in production!")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
		if err != nil {
			log.Warn().Msg("An invalid log level was provided. Using the 'info' fallback value.")
			logLevel = zerolog.InfoLevel
		}
		zerolog.SetGlobalLevel(logLevel)
	}

	// Start the webhook server
	wh := &webhook.Server{
		Address:      cfg.ListenAddress,
		ErrorMessage: cfg.ErrorMessage,
	}
	log.Info().Str("address", cfg.ListenAddress).Msg("Starting the HTTP server...")
	go func() {
		if err := wh.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Could not start the HTTP server.")
		}
	}()
	defer func() {
		log.Info().Msg("Shutting down the HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := wh.Stop(ctx); err != nil {
			log.Err(err).Msg("Could not gracefully shut down the HTTP server.")
		}
	}()

	// Wait for a Ctrl-C signal
	log.Info().Msg("The application has been started. To stop it press Ctrl-C.")
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
}
