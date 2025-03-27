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
	"github.com/Impisigmatus/service_core/postgres"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/autogen/server"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/cameras/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/LeonKote/PSSVTelegramBot/microservices/cameras/autogen/docs"
)

// @title Cameras API
// @version 1.0
// @description %README_FILE%
// @host localhost:8000
// @BasePath /api
func main() {
	ctx := context.Background()
	log.Init(log.LevelDebug)
	cfg := config.MakeConfig()
	repo := sqlx.NewDb(
		postgres.NewPostgres(
			postgres.Config{
				Hostname: cfg.PgHost,
				Port:     cfg.PgPort,
				Database: cfg.PgDB,
				User:     cfg.PgUser,
				Password: cfg.PgPass,
			},
		), "pgx")

	app := service.MakeApplication(ctx, repo, cfg)

	router := getRouter(ctx, app, cfg)
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

func getRouter(ctx context.Context, app *service.Application, cfg config.Config) *chi.Mux {
	transport := service.NewTransport(ctx, app)

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
			//middlewares.Authorization([]string{cfg.BasicLogin, cfg.BasicPass}),
			//),
			middlewares.Logger(),
		),
	)

	return router
}
