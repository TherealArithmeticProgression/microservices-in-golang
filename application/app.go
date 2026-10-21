package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
	config config
}

func New() *App {
	cfg := loadconfig()
	app := &App{
		config: cfg,
		rdb:    redis.NewClient(&redis.Options{Addr: cfg.redis_address}),
	}
	app.loadRoutes()
	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.server_port),
		Handler: a.router,
	}
	if err := a.rdb.Ping().Err(); err != nil {
		return fmt.Errorf("Failed to connect to redis: %w", err)
	}
	defer func() { //defer is the last function you look into, before leaving the current function.
		if err := a.rdb.Close(); err != nil {
			fmt.Println("Failed to close your redis. Look into it.", err)
		}
	}()

	fmt.Println("Starting the server.")
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("Failed to start the server: %w", err)
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel() //close redis when you close the door/context
		if err := server.Shutdown(timeout); err != nil {
			return fmt.Errorf("Failed to shut down the server: %w", err)
		}
		return nil
	}
}
