package application

import (
	"context"
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

func New(config config) *App {
	app := &App{
		router: loadRoutes(),
		rdb:    redis.NewClient(&redis.Options{}),
	}
	app.loadRoutes()
	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.server_port),
		Handler: a.router,
	}
	err := a.rdb.Ping().Err()
	if err != nil {
		return fmt.Errorf("Failed to connect to redis: %w", err)
	}
	defer func() { //defer is the last function you look into, before leaving the current function.
		if err := a.rdb.Close(); err != nil {
			fmt.Println("Failed to close your redis. Look into it.", err)
		}
	}()
	err = server.ListenAndServe()
	fmt.Println("Starting the server.")
	ch := make(chan error, 1)
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("Failed to start the server: %w", err)
		}
		close(ch)
	}()
	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel() //close redis when you close the door/context
		return server.Shutdown(timeout)

	}

}
