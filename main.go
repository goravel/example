package main

import (
	"github.com/goravel/framework/facades"

	"goravel/bootstrap"
)

func main() {
	// This bootstraps the framework and gets it ready for use.
	bootstrap.Boot()

	// Start HTTP server by facades.Route().
	go func() {
		if err := facades.Route().Run(); err != nil {
			facades.Log().Errorf("Route run error: %v", err)
		}
	}()

	// Start GRPC server
	go func() {
		if err := facades.Grpc().Run(); err != nil {
			facades.Log().Errorf("Run grpc error: %+v", err)
		}
	}()

	select {}
}
