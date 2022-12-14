package app

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/ssssargsian/uniplay/internal/config"
	v1 "github.com/ssssargsian/uniplay/internal/handler/v1"
	"github.com/ssssargsian/uniplay/internal/pkg/httpserver"
	"github.com/ssssargsian/uniplay/internal/pkg/logger"
	"github.com/ssssargsian/uniplay/internal/pkg/pgclient"
	"github.com/ssssargsian/uniplay/internal/postgres"
	"github.com/ssssargsian/uniplay/internal/service"

	v1gen "github.com/ssssargsian/uniplay/internal/gen/oapi/v1"
)

func Run(conf *config.Config) {
	l, err := logger.New(os.Stderr, conf.Log.Level)
	if err != nil {
		log.Fatalf("logger.New: %s", err.Error())
	}

	// db
	pgClient, err := pgclient.New(conf.PG.URL, pgclient.WithMaxConns(conf.PG.MaxConns))
	if err != nil {
		l.Fatal("pgclient.New", zap.Error(err))
	}

	// repos
	replayRepo := postgres.NewReplayRepo(l, pgClient)
	playerRepo := postgres.NewPlayerRepo(l, pgClient)
	compendiumRepo := postgres.NewCompendiumRepo(l, pgClient)

	// services
	replayService := service.NewReplay(l, replayRepo)
	playerService := service.NewPlayer(playerRepo)
	compendiumService := service.NewCompendium(compendiumRepo)

	// init handlers
	mux := chi.NewMux()
	mux.Use(middleware.Logger, middleware.Recoverer)

	handlerV1 := v1.NewHandler(l, replayService, playerService, compendiumService)
	v1gen.HandlerFromMuxWithBaseURL(handlerV1, mux, "/v1")

	runHTTPServer(mux, l, conf.HTTP.Port)
}

func runHTTPServer(mux http.Handler, l *zap.Logger, port string) {
	l.Info("starting http server", zap.String("port", port))

	httpServer := httpserver.New(mux, httpserver.WithPort(port))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("received signal from httpserver", zap.String("signal", s.String()))
	case err := <-httpServer.Notify():
		l.Info("got error from http server notify", zap.Error(err))
	}

	if err := httpServer.Shutdown(); err != nil {
		l.Info("got error on http server shutdown", zap.Error(err))
	}
}
