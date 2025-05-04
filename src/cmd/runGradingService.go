package main

import (
	"context"
	"fmt"
	"os"
	"pluralsight-go-building-distributed-apps/grades"
	"pluralsight-go-building-distributed-apps/log"
	"pluralsight-go-building-distributed-apps/pkg/util"
	"pluralsight-go-building-distributed-apps/registry"
	"pluralsight-go-building-distributed-apps/service"

	stLog "log"
)

func main() {
	port := util.StringOr(os.Getenv("GRADING_SERVICE_PORT"), "6000")
	host := util.StringOr(os.Getenv("SERVICE_HOSTNAME"), "localhost")

	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	registration := registry.Registration{
		ServiceName:      registry.GradingService,
		ServiceURL:       serviceAddress,
		RequiredServices: []registry.ServiceName{registry.LogService},
		ServiceUpdateURL: fmt.Sprintf("%s/services", serviceAddress),
	}

	ctx, err := service.Start(context.Background(), host, port, registration, grades.RegisterHandlers)

	if err != nil {
		stLog.Fatal(err)
	}

	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		fmt.Printf("logging service found at: %v\n", logProvider)
		log.SetClientLogger(logProvider, registration.ServiceName)
	}

	<-ctx.Done()
	fmt.Println("Shutting down grading service")
}
