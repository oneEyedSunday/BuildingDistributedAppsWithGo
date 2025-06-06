package main

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

/*
We are not reusing the service.go because it's eventually going to have functionality
specifically designed to handle client services. So the registry service itself won't be able
to take advantage of that.
*/
func main() {
	registry.SetupRegistryService()
	http.Handle("/services", &registry.RegistryService{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	var srv http.Server
	srv.Addr = fmt.Sprint(":", registry.ServerPort)

	// 1) goroutine to start the server
	go func() {
		fmt.Printf("Registry service started on http://%s. Press an key to stop. \n", srv.Addr)
		log.Println(srv.ListenAndServe())

		// if ListenAndServe() returns, it means that an error has occurred, so we need to cancel the context.
		cancel()
	}()

	// 2) watch for signal cancellation from os (SIGINT, SIGTERM etc)
	go func() {
		<-sig

		srv.Shutdown(ctx)
		cancel()
	}()

	<-ctx.Done()

	fmt.Println("Shutting down registry service")
}
