package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"hoel-app/backend/internal/config"
	"hoel-app/backend/internal/db"
	"hoel-app/backend/internal/integration"
	"hoel-app/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.OpenSQLite(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			log.Printf("database close error: %v", closeErr)
		}
	}()

	if err := db.ApplyMigrations(database, cfg.MigrationsDir); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	monitoringRepository := db.NewMonitoringRepository(database)
	pghRepository := db.NewPGHRepository(database)
	tickTickRepository := db.NewTickTickRepository(database)
	tickTickOAuthRepository := db.NewTickTickOAuthRepository(database)
	integrationClient := integration.NewClient(
		cfg.OutboundTimeout,
		cfg.OutboundRetries,
		cfg.OutboundBackoff,
		monitoringRepository,
	)
	tickTickOAuth := integration.NewTickTickOAuthService(
		integrationClient,
		tickTickOAuthRepository,
		cfg.TickTickAuthURL,
		cfg.TickTickTokenURL,
		cfg.TickTickClientID,
		cfg.TickTickClientSecret,
		cfg.TickTickRedirectURI,
		cfg.TickTickToken,
	)
	pghService := integration.NewPGHService(integrationClient, pghRepository, cfg.PGHEndpoint, cfg.PGHPollInterval)
	tickTickService := integration.NewTickTickService(
		integrationClient,
		tickTickRepository,
		tickTickOAuth,
		cfg.TickTickAPIRoot,
		cfg.TickTickProject,
		cfg.TickTickPoll,
	)

	runtimeContext, runtimeCancel := context.WithCancel(context.Background())
	defer runtimeCancel()

	if pghService.Enabled() {
		go pghService.Start(runtimeContext)
	}

	if tickTickService.Enabled() {
		go tickTickService.Start(runtimeContext)
	}

	apiServer := server.New(
		cfg.Address(),
		cfg.ReadTimeout,
		cfg.WriteTimeout,
		cfg.AllowedOrigins,
		monitoringRepository,
		pghRepository,
		tickTickRepository,
		tickTickService,
		integrationClient,
		tickTickOAuth,
		cfg.TickTickAPIRoot,
	)
	errChannel := make(chan error, 1)

	go func() {
		errChannel <- apiServer.ListenAndServe()
	}()

	log.Printf("server listening on %s", cfg.Address())

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stopSignal:
		log.Printf("received signal %s, shutting down", sig.String())
	case serverErr := <-errChannel:
		if !errors.Is(serverErr, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", serverErr)
		}
	}

	runtimeCancel()

	shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := apiServer.Shutdown(shutdownContext); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}

	fmt.Println("server stopped cleanly")
}
