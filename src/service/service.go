package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pluralsight-go-building-distributed-apps/registry"
	"syscall"
)

func Start(ctx context.Context, host, port string, reg registry.Registration, registerHandleFn func()) (context.Context, error) {

	registerHandleFn()
	ctx = start(ctx, reg, host, port)

	// register service
	if err := registry.RegisterService(reg); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func start(ctx context.Context, reg registry.Registration, host, port string) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	var svc http.Server
	svc.Addr = ":" + port

	go func() {
		// start up server
		log.Println(svc.ListenAndServe())

		// an error occured
		// cleanup
		cancel()
	}()

	go func() {
		fmt.Printf("%v started on http://%s:%s. Press an key to stop. \n", reg.ServiceName, host, port)
		<-sig

		if err := registry.ShutdownService(reg.ServiceURL); err != nil {
			log.Println(err)
		}
		svc.Shutdown(ctx)
		cancel()
	}()

	return ctx
}
