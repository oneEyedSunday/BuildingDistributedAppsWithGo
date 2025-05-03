package main

import (
	"context"
	"fmt"
	"os"
	"pluralsight-go-building-distributed-apps/log"
	"pluralsight-go-building-distributed-apps/pkg/util"
	"pluralsight-go-building-distributed-apps/registry"
	"pluralsight-go-building-distributed-apps/service"

	stLog "log"
)

func main() {
	log.Run("app.log")

	port := util.StringOr(os.Getenv("LOG_SERVICE_PORT"), "4000")
	host := "localhost"

	ctx, err := service.Start(context.Background(), host, port, registry.Registration{
		ServiceName: registry.LogService,
		ServiceURL:  fmt.Sprintf("http://%s:%s", host, port),
	}, log.RegisterHandlers)

	if err != nil {
		stLog.Fatal(err)
	}

	<-ctx.Done()
	fmt.Println("Shutting down log service")
}
