package main

import (
	"context"
	"fmt"
	"microservices-in-golang/application"
)

func main() {
	app := application.New()
	err := app.Start(context.TODO())
	fmt.Println("This is working.")
	if err != nil {
		fmt.Println("Got an error while starting the application ", err)
	}
}
