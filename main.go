package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"

	"goravel/bootstrap"
)

func main() {
	// This bootstraps the framework and gets it ready for use.
	bootstrap.Boot()

	// Create a channel to listen for OS signals
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start http server by facades.Route().
	go func() {
		if err := facades.Route().Run(); err != nil {
			facades.Log().Errorf("Route run error: %v", err)
		}
	}()

	// Listen for the OS signal
	go func() {
		<-quit
		if err := facades.Route().Shutdown(); err != nil {
			facades.Log().Errorf("Route Rhutdown error: %v", err)
		}

		os.Exit(0)
	}()

	// Start grpc server by facades.Grpc().
	go func() {
		if err := facades.Grpc().Run(); err != nil {
			facades.Log().Errorf("Grpc run error: %v", err)
		}
	}()
	go func() {
		//测试队列
		queueErr := facades.Queue().Worker(&queue.Args{
			Queue:      "test_job", //队列名称
			Concurrent: 4,          //4个并发
		}).Run()
		if queueErr != nil {
			facades.Log().Errorf("Queue run error: %v", queueErr)
		}
	}()
	select {}
}
