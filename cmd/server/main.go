package main

import (
	"context"
	"fmt"
	"github.com/ZertGraf/avito-test/internal/bootstrap"

	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// initialize application
	app, err := bootstrap.New()
	if err != nil {
		fmt.Printf("failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// establish external connections
	if err = app.Init(ctx); err != nil {
		app.Logger.Error("failed to establish connections", "error", err)
		os.Exit(1)
	}

	// setup graceful shutdown handling
	setupGracefulShutdown(ctx, cancel, app)

	// start the file service
	app.Logger.Info("starting avito-test service",
		"version", "0.1.0",
		"environment", app.Config.Environment,
		"log_level", app.Config.LogLevel)

	app.Logger.Info("service started successfully")
	// wait for shutdown signal
	<-ctx.Done()
	app.Logger.Info("received shutdown signal, initiating graceful shutdown")

	// graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// shutdown application components
	if err = app.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error("application shutdown failed", "error", err)
		os.Exit(1)
	}

	app.Logger.Info("service stopped gracefully")

}

// setupGracefulShutdown configures signal handling for clean shutdown
func setupGracefulShutdown(ctx context.Context, cancel context.CancelFunc, app *bootstrap.Application) {
	// channel for receiving os signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigChan:
			app.Logger.Info("received shutdown signal", "signal", sig.String())
			// cancel main context to initiate shutdown
			cancel()
		case <-ctx.Done():
			// context already cancelled
		}
	}()
}
