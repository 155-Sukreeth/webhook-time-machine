package app

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/155-Sukreeth/webhook-time-machine/internal/api"
	"github.com/155-Sukreeth/webhook-time-machine/internal/config"
	"github.com/155-Sukreeth/webhook-time-machine/internal/dashboard"
	"github.com/155-Sukreeth/webhook-time-machine/internal/logger"
	"github.com/155-Sukreeth/webhook-time-machine/internal/models"
	"github.com/155-Sukreeth/webhook-time-machine/internal/proxy"
	"github.com/155-Sukreeth/webhook-time-machine/internal/replay"
	"github.com/155-Sukreeth/webhook-time-machine/internal/storage"
)

type App struct {
	cfg   *models.Config
	log   *logger.Logger
	webFS embed.FS
}

func New(cfg *models.Config, webFS embed.FS) *App {
	return &App{
		cfg:   cfg,
		log:   logger.New(),
		webFS: webFS,
	}
}

func (a *App) Run(ctx context.Context) error {
	store, err := storage.New(a.cfg.DBPath)
	if err != nil {
		return fmt.Errorf("storage error: %w", err)
	}
	defer store.Close()

	if err := store.InitSchema(ctx); err != nil {
		return fmt.Errorf("schema init error: %w", err)
	}

	executor := replay.NewExecutor(store)
	proxySrv, err := proxy.NewServer(a.cfg.ForwardURL, store)
	if err != nil {
		return fmt.Errorf("proxy init error: %w", err)
	}

	apiH := api.NewAPIHandler(store, executor, a.cfg)
	dashSrv := dashboard.NewServer(apiH, a.webFS)

	proxyHTTP := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.cfg.Port),
		Handler:           proxySrv,
		ReadHeaderTimeout: 5 * time.Second,
	}

	dashHTTP := &http.Server{
		Addr:              fmt.Sprintf(":%d", a.cfg.UIPort),
		Handler:           dashSrv.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	a.log.Info("================================================================")
	a.log.Info(" Local Webhook Time Machine (wtm)")
	a.log.Info(" Proxy Listening : http://localhost:%d", a.cfg.Port)
	a.log.Info(" Forward Target  : %s", a.cfg.ForwardURL)
	a.log.Info(" Dashboard UI    : http://localhost:%d", a.cfg.UIPort)
	a.log.Info(" Database Storage: %s", a.cfg.DBPath)
	a.log.Info("================================================================")

	errChan := make(chan error, 2)
	go func() {
		if err := proxyHTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("proxy server error: %w", err)
		}
	}()

	go func() {
		if err := dashHTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("dashboard server error: %w", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case sig := <-sigChan:
		a.log.Info("Received signal %s, initiating graceful shutdown...", sig)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = proxyHTTP.Shutdown(shutdownCtx)
		_ = dashHTTP.Shutdown(shutdownCtx)
		a.log.Info("Clean shutdown complete.")
		return nil
	}
}

func InitConfig(targetPath string) error {
	return config.WriteDefaultConfigFile(targetPath)
}
