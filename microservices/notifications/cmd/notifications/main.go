package main

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/Impisigmatus/service_core/log"
	"github.com/Impisigmatus/service_core/middlewares"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/autogen/server"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/bot"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/notifications/internal/service"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/LeonKote/PSSVTelegramBot/microservices/notifications/autogen/docs"
)

func main() {
	logger := log.New(log.LevelDebug)
	cfg := config.MakeConfig(logger)

	bot := bot.NewBot(logger, *cfg)

	router := getRouter(bot, *cfg)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Panic().Msgf("Invalid service starting: %s", err)
		}
		logger.Info().Msg("Service stopped")
	}()
	logger.Info().Msg("Service started")

	channel := make(chan os.Signal, 1)
	signal.Notify(channel,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	<-channel

	if err := server.Shutdown(context.Background()); err != nil {
		logger.Panic().Msgf("Invalid service stopping: %s", err)
	}
}

func getRouter(bot *bot.Bot, cfg config.Config) *chi.Mux {
	transport := service.NewTransport(bot)

	router := chi.NewRouter()
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	router.Handle("/api/*",
		middlewares.Use(
			middlewares.Use(
				middlewares.Use(
					server.Handler(transport),
					middlewares.Authorization([]string{cfg.BasicAuth}),
				),
				middlewares.ContextLogger()),
			middlewares.RequestID(cfg.Logger),
		),
	)

	return router
}
