package main

import (
	"github.com/devtron-labs/chart-sync/internals"
	"github.com/devtron-labs/chart-sync/pkg"
	"github.com/go-pg/pg"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"strconv"
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
	go func() {
		// Then modify the line to:
		err := http.ListenAndServe(":"+strconv.Itoa(app.configuration.PrometheusMatrixPort), nil)
		if err != nil {
			app.Logger.Errorw("error in starting prometheus server", "err", err)
		}
	}()

	// Track overall sync time
	start := time.Now()
	// Start the sync service
	_, err := app.syncService.Sync()
	if err != nil {
		app.Logger.Errorw("err", "err", err)
		internals.RepoSyncDuration.WithLabelValues("all", "all", err.Error()).Observe(time.Since(start).Seconds())
	} else {
		internals.RepoSyncDuration.WithLabelValues("all", "all", "").Observe(time.Since(start).Seconds())
	}

	// sleep for ShutDownInterval seconds to give time for prometheus to scrape the metrics
	time.Sleep(time.Duration(app.configuration.AppSyncJobShutDownWaitDuration) * time.Second)
}
