package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"test-builder/internal/auth"
	"test-builder/internal/config"
	"test-builder/internal/database"
	"test-builder/internal/httpserver"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	cfg := config.MustLoad()

	poolCtx, poolCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer poolCancel()

	pool, err := database.NewPool(poolCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	log.Printf("connected to PostgreSQL")

	
	authRepo := auth.NewRepository(pool)

	authService, err := auth.NewService(authRepo)
	if err != nil {
		return err
	}

	
	authHandler := auth.NewHandler(authService, cfg.AppEnv)

	// HTTP-роутер
	router := httpserver.New(authHandler, cfg.FrontendOrigin, cfg.AppEnv)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("starting server on %s (env=%s)", cfg.HTTPAddr, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Ожидание сигнала ИЛИ падения сервера
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	select {
	case err := <-serverErr:
		return err
	case sig := <-stop:
		log.Printf("received signal %s, shutting down...", sig)
	}

	//  Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	log.Println("stopped")
	return nil
}