package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/elangreza/content-management-system/cmd/server/config"
	"github.com/elangreza/content-management-system/internal/postgresql"
	"github.com/elangreza/content-management-system/internal/rest"
	"github.com/elangreza/content-management-system/internal/service"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.LoadConfig()
	errChecker(err)

	dn, err := config.SetupDB(cfg)
	errChecker(err)

	// deps, err := InitializeProductHandler(cfg)
	// errChecker(err)

	c := chi.NewRouter()

	// repositories
	ur := postgresql.NewUserRepo(dn)
	tr := postgresql.NewTokenRepo(dn)
	ar := postgresql.NewArticleRepo(dn)

	// services
	authService := service.NewAuthService(ur, tr)
	profileService := service.NewProfileService(ur)
	articleService := service.NewArticleService(ar)

	rest.NewAuthRouter(c, authService)
	// private routes
	rest.NewHandlerWithMiddleware(c, profileService, authService, articleService)

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.HTTP_PORT),
		Handler:        c,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	slog.Info("server started", "port", cfg.HTTP_PORT)

	<-gracefulShutdown(context.Background(), 5*time.Second,
		operation{
			name: "server",
			shutdownFunc: func(ctx context.Context) error {
				return srv.Shutdown(ctx)
			}},
		operation{
			name: "postgres",
			shutdownFunc: func(ctx context.Context) error {
				return dn.Close()
			}},
	)
}

func errChecker(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type operation struct {
	name         string
	shutdownFunc func(ctx context.Context) error
}

func gracefulShutdown(ctx context.Context, timeout time.Duration, ops ...operation) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)

		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-s

		slog.Info("shutting down")

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		go func() {
			<-ctx.Done()
			slog.Info("force quit the app")
			wait <- struct{}{}
		}()

		var wg sync.WaitGroup

		for key, op := range ops {
			wg.Add(1)
			go func(key int, op operation) {
				defer wg.Done()

				slog.Info(op.name, "shutdown", "started")

				if err := op.shutdownFunc(ctx); err != nil {
					slog.Error(op.name, "err", err.Error())
					return
				}

				slog.Info(op.name, "shutdown", "finished")
			}(key, op)
		}

		wg.Wait()
	}()

	return wait
}
