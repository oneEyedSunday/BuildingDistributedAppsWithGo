package main

import (
	"context"
	"fmt"
	stLog "log"
	stlog "log"
	"os"
	"pluralsight-go-building-distributed-apps/log"
	"pluralsight-go-building-distributed-apps/pkg/util"
	"pluralsight-go-building-distributed-apps/registry"
	"pluralsight-go-building-distributed-apps/service"
	"pluralsight-go-building-distributed-apps/teacherportal"
)

func main() {

	port := util.StringOr(os.Getenv("TEACHER_PORTAL_PORT"), "5500")
	host := util.StringOr(os.Getenv("SERVICE_HOSTNAME"), "localhost")

	err := teacherportal.ImportTemplates()
	if err != nil {
		stlog.Fatal(err)
	}

	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	registration := registry.Registration{
		ServiceName:      registry.TeacherPortal,
		ServiceURL:       serviceAddress,
		RequiredServices: []registry.ServiceName{registry.LogService, registry.GradingService},
		ServiceUpdateURL: fmt.Sprintf("%s/services", serviceAddress),
		HeartbeatURL:     fmt.Sprintf("%s/heartbeat", serviceAddress),
	}

	ctx, err := service.Start(context.Background(), host, port, registration, teacherportal.RegisterHandlers)

	if err != nil {
		stLog.Fatal(err)
	}

	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		fmt.Printf("logging service found at: %v\n", logProvider)
		log.SetClientLogger(logProvider, registration.ServiceName)
	}

	<-ctx.Done()
	fmt.Println("Shutting down teacher portal")
}
