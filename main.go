package main

import (
	"context"
	"fmt"
	"microservices-in-golang/application"
	"os"
	"os/signal"
)

func main() {
	app := application.New()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt) //root context - should be only created in main and init
	defer cancel()                                                          //last function you should call (once the )
	err := app.Start(ctx)
	fmt.Println("This is working.")
	if err != nil {
		fmt.Println("Got an error while starting the application ", err)
	}
}
