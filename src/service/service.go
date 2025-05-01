package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func Start(ctx context.Context, serviceName, host, port string, registerHandleFn func()) (context.Context, error) {

	registerHandleFn()
	ctx = start(ctx, serviceName, host, port)

	return ctx, nil
}

func start(ctx context.Context, serviceName, host, port string) context.Context {
	ctx, cancel := context.WithCancel(ctx)
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

		fmt.Printf("%v started on http://%s:%s. Press an key to stop. \n", serviceName, host, port)
		var s string
		fmt.Scanln(&s)

		// shutdown server on key input???
		svc.Shutdown(ctx)
		cancel()

	}()

	return ctx
}
