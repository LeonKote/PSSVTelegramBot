package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/Impisigmatus/service_core/log"
	"github.com/LeonKote/PSSVTelegramBot/microservices/openCV/internal/api"
	"github.com/LeonKote/PSSVTelegramBot/microservices/openCV/internal/app"
	"github.com/LeonKote/PSSVTelegramBot/microservices/openCV/internal/config"
	"github.com/go-chi/chi/v5"
)

func main() {
	logger := log.New(log.LevelDebug)
	ctx := context.Background()
	cfg := config.MakeConfig(logger)

	api := api.NewCameraApi(cfg)
	cameras, err := api.GetAllCameras()
	if err != nil {
		logger.Panic().Msgf("Invalid service starting: %s", err)
	}

	apps := make(map[string]*app.Application)
	for _, camera := range cameras {
		url := fmt.Sprintf("%s/%s", cfg.StreamUrl, camera.Name)
		app := app.MakeApplication(logger, url, cfg)

		go func() {
			go app.CheckPhoto(logger, ctx)
		}()

		apps[camera.Name] = app
	}

	router := getRouter(cfg)
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

func getRouter(cfg config.Config) *chi.Mux {
	router := chi.NewRouter()

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return router
}
