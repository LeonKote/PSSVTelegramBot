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
	"github.com/LeonKote/PSSVTelegramBot/microservices/users/autogen/server"
	"github.com/LeonKote/PSSVTelegramBot/microservices/users/internal/config"
	"github.com/LeonKote/PSSVTelegramBot/microservices/users/internal/repository"
	"github.com/LeonKote/PSSVTelegramBot/microservices/users/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/LeonKote/PSSVTelegramBot/microservices/users/autogen/docs"
)

// @title Users API
// @version 3.0
// @description %README_FILE%
// @host localhost:8000
// @BasePath /api
func main() {
	cfg := config.MakeConfig()
	log.Init(log.LevelDebug)

	repo := repository.NewUsersRepository(
		sqlx.NewDb(
			postgres.NewPostgres(
				postgres.Config{
					Hostname: cfg.PgHost,
					Port:     cfg.PgPort,
					Database: cfg.PgDB,
					User:     cfg.PgUser,
					Password: cfg.PgPass,
				},
			), "pgx"),
	)

	router := getRouter(repo, cfg)

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

func getRouter(repo repository.IUsersRepository, cfg config.Config) *chi.Mux {
	transport := service.NewTransport(repo)

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
