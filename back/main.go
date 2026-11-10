// @title           Test Builder Platform API
// @version         1.0
// @description     API для платформы создания тестов
// @host            localhost
// @BasePath		 /
// @securityDefinitions.apikey  CookieAuth
// @in                          header
// @name                        Cookie
// @description                 Вставьте сюда строку вида: tbp_session=<токен из cookie>

// @securityDefinitions.apikey  CsrfToken
// @in                          header
// @name                        X-CSRF-Token
// @description                 CSRF-токен (из ответа POST /auth/login или GET /auth/me)
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
	"time"

	"github.com/redis/go-redis/v9"

	"test-builder/internal/auth"
	"test-builder/internal/config"
	"test-builder/internal/dashboard"
	"test-builder/internal/database"
	"test-builder/internal/httpserver"
	"test-builder/internal/publications"
	"test-builder/internal/tests"
	"test-builder/internal/attempts"
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
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   0,
	})
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}
	defer rdb.Close()
	log.Printf("connected to Redis at %s", cfg.RedisAddr)

	authRepo := auth.NewRepository(pool)

	authService, err := auth.NewService(authRepo)
	if err != nil {
		return err
	}

	authHandler := auth.NewHandler(authService, cfg.AppEnv)
	dashboardRepo := dashboard.NewRepository(pool)
	dashboardSvc := dashboard.NewService(dashboardRepo)
	dashboardHandler := dashboard.NewHandler(dashboardSvc)
	testsRepo := tests.NewRepository(pool)
	testsSvc := tests.NewService(testsRepo)
	testsHandler := tests.NewHandler(testsSvc)
	publicationsRepo := publications.NewRepository(pool)
	publicationsService := publications.NewService(publicationsRepo, pool, cfg.PublicBaseURL)
	publicationsHandler := publications.NewHandler(publicationsService)
	attemptsRepo := attempts.NewRepository(pool)
	attemptsSvc := attempts.NewService(attemptsRepo, pool)
	attemptsHandler := attempts.NewHandler(attemptsSvc, cfg.AppEnv)
	// HTTP-роутер
	router := httpserver.New(authHandler, dashboardHandler, testsHandler, publicationsHandler, attemptsHandler, rdb, cfg.FrontendOrigin, cfg.AppEnv)

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
