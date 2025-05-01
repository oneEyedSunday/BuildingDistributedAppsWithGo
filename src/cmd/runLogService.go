package main

import (
	"context"
	"fmt"
	"pluralsight-go-building-distributed-apps/log"
	"pluralsight-go-building-distributed-apps/service"

	stLog "log"
)

func main() {
	log.Run("../dist/app.log")

	// TODO from config
	host, port := "localhost", "4000"

	ctx, err := service.Start(context.Background(), "Log Service", host, port, log.RegisterHandlers)

	if err != nil {
		stLog.Fatal(err)
	}

	<-ctx.Done()
	fmt.Println("Shutting down log service")
}
