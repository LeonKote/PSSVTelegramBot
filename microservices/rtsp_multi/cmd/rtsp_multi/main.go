package main

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/Impisigmatus/service_core/log"
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
	cfg := config.MakeConfig()
	log.Init(log.LevelDebug)

	api := api.NewCameraApi(cfg)
	cameras, err := api.GetAllCameras()
	if err != nil {
		log.Panicf("Invalid service starting: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apps := make(map[string]*service.Application)
	for _, camera := range cameras {
		log.Debugf("Camera: %s", camera.Name)
		// создаём копию объекта camera, чтобы избежать гонки
		cam := camera

		app := service.NewApp(cam.Rtsp)
		go func(name string, app *service.Application) {
			log.Infof("Launching camera: %s", name)
			go func() {
				log.Info("FFmpeg started")
				err := app.Run(ctx)
				if err != nil {
					log.Errorf("FFmpeg error: %v", err)
				}
				log.Info("FFmpeg stopped")
			}()
			go func() {
				log.Info("DistributeStream started")
				err := app.DistributeStream(ctx)
				if err != nil {
					log.Errorf("DistributeStream error: %v", err)
				}
				log.Info("DistributeStream stopped")
			}()
		}(cam.Name, app)

		apps[cam.Name] = app
	}

	router := getRouter(apps)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	// Запуск HTTP-сервера
	go func() {
		log.Infof("Service listening on %s", cfg.Address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panicf("HTTP server error: %s", err)
		}
		log.Info("HTTP server stopped")
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

	sig := <-stop
	log.Infof("Received signal: %v. Shutting down...", sig)

	// Завершение всех потоков
	cancel()

	// Корректное завершение HTTP-сервера
	if err := server.Shutdown(ctx); err != nil {
		log.Panicf("Error shutting down server: %s", err)
	}

	log.Info("Service gracefully stopped")
}

func getRouter(apps map[string]*service.Application) *chi.Mux {
	router := chi.NewRouter()

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	router.Route("/stream", func(r chi.Router) {
		r.Get("/{camera_name}", func(w http.ResponseWriter, r *http.Request) {
			cameraName := chi.URLParam(r, "camera_name")
			log.Debugf("Stream request for camera: %s.", cameraName)
			if app, ok := apps[cameraName]; ok {
				app.StreamHandler(w, r)
			} else {
				http.Error(w, "Camera not found", http.StatusNotFound)
			}
		})
	})

	return router
}
