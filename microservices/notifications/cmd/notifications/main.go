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
	cfg := config.MakeConfig()
	log.Init(log.LevelDebug)

	transport := service.NewTransport(
		bot.NewBot(*cfg),
	)

	router := getRouter(context.Background(), transport, cfg.BasicLogin, cfg.BasicPass)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panicf("Invalid service starting: %s", err)
		}
		log.Info("Service stopped")
	}()
	log.Info("Service started")

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
		log.Panicf("Invalid service stopping: %s", err)
	}
}

func getRouter(ctx context.Context, transport server.ServerInterface, basicLogin string, basicPass string) *chi.Mux {
	router := chi.NewRouter()
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	router.Handle("/api/*",
		middlewares.Use(
			// middlewares.Use(
			server.Handler(transport),
			//middlewares.Authorization([]string{basicLogin, basicPass}),
			//),
			middlewares.Logger(),
		),
	)

	return router
}
