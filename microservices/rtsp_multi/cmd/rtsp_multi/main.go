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
	"github.com/LeonKote/PSSVTelegramBot/microservices/rtsp_multi/internal/api"
	"github.com/LeonKote/PSSVTelegramBot/microservices/rtsp_multi/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/rtsp_multi/internal/service"
	"github.com/go-chi/chi/v5"
)

// @title Cameras API
// @version 1.0
// @description %README_FILE%
// @host localhost:8000
// @BasePath /api
func main() {
	logger := log.New(log.LevelDebug)
	cfg := config.MakeConfig(logger)

	api := api.NewCameraApi(cfg)
	cameras, err := api.GetAllCameras()
	if err != nil {
		logger.Panic().Msgf("Invalid service starting: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apps := make(map[string]*service.Application)
	for _, camera := range cameras {
		logger.Info().Msgf("Camera: %s", camera.Name)
		// создаём копию объекта camera, чтобы избежать гонки
		cam := camera

		app := service.NewApp(cam.Rtsp)
		go func(name string, app *service.Application) {
			logger.Info().Msgf("Launching camera: %s", name)
			go func() {
				logger.Info().Msg("FFmpeg started")
				err := app.Run(logger, ctx)
				if err != nil {
					logger.Error().Msgf("FFmpeg error: %v", err)
				}
				logger.Info().Msg("FFmpeg stopped")
			}()
			go func() {
				logger.Info().Msg("DistributeStream started")
				err := app.DistributeStream(logger, ctx)
				if err != nil {
					logger.Error().Msgf("DistributeStream error: %v", err)
				}
				logger.Info().Msg("DistributeStream stopped")
			}()
		}(cam.Name, app)

		apps[cam.Name] = app
	}

	router := getRouter(apps, cfg)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	// Запуск HTTP-сервера
	go func() {
		logger.Info().Msgf("Service listening on %s", cfg.Address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Panic().Msgf("HTTP server error: %s", err)
		}
		logger.Info().Msg("HTTP server stopped")
	}()

	// Обработка сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	<-stop

	// Завершение всех потоков
	cancel()

	// Корректное завершение HTTP-сервера
	if err := server.Shutdown(ctx); err != nil {
		logger.Panic().Msgf("Error shutting down server: %s", err)
	}

	logger.Info().Msg("Service gracefully stopped")
}

func getRouter(apps map[string]*service.Application, cfg config.Config) *chi.Mux {
	router := chi.NewRouter()

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	router.Route("/stream", func(r chi.Router) {
		r.Handle("/{camera_name}",
			middlewares.Use(
				middlewares.Use(
					middlewares.Use(
						service.NewTransport(apps),
						middlewares.Authorization([]string{cfg.BasicAuth}),
					),
					middlewares.ContextLogger()),
				middlewares.RequestID(cfg.Logger),
			),
		)
	})

	return router
}
