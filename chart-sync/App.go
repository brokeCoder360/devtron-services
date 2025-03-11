package main

import (
	"github.com/devtron-labs/chart-sync/internals"
	"github.com/devtron-labs/chart-sync/pkg"
	"github.com/go-pg/pg"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type App struct {
	Logger        *zap.SugaredLogger
	db            *pg.DB
	syncService   pkg.SyncService
	configuration *internals.Configuration
}

func NewApp(Logger *zap.SugaredLogger,
	db *pg.DB,
	syncService pkg.SyncService,
	configuration *internals.Configuration) *App {
	return &App{
		Logger:        Logger,
		db:            db,
		syncService:   syncService,
		configuration: configuration,
	}
}

func (app *App) Start() {
	// Set up the /metrics endpoint for Prometheus to scrape
	http.Handle("/metrics", promhttp.Handler())

	// Start the sync service
	_, err := app.syncService.Sync()
	// sleep for ShutDownInterval seconds to give time for prometheus to scrape the metrics
	time.Sleep(time.Duration(app.configuration.ShutDownInterval) * time.Second)

	if err != nil {
		app.Logger.Errorw("err", "err", err)
	}
}
