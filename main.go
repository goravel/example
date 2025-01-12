package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/tableflip"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/env"

	"goravel/bootstrap"
)

func main() {
	// This bootstraps the framework and gets it ready for use
	bootstrap.Boot()

	// Start grpc server by facades.Grpc()
	go func() {
		if err := facades.Grpc().Run(); err != nil {
			facades.Log().Errorf("Grpc run error: %v", err)
		}
	}()

	// Start queue server by facades.Queue()
	go func() {
		if err := facades.Queue().Worker().Run(); err != nil {
			facades.Log().Errorf("Queue run error: %v", err)
		}
	}()

	// Start http server by facades.Route()
	if !env.IsWindows() {
		if err := runGraceful(); err != nil {
			facades.Log().Errorf("Route run error: %v", err)
		}
	} else {
		if err := facades.Route().Run(); err != nil {
			facades.Log().Errorf("Route run error: %v", err)
		}
	}
}

// runGraceful graceful start http server
func runGraceful() error {
	upg, err := tableflip.New(tableflip.Options{})
	if err != nil {
		return err
	}
	defer upg.Stop()

	// Listen for the process signal to trigger the tableflip upgrade.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for range sig {
			if err = upg.Upgrade(); err != nil {
				facades.Log().Errorf("Graceful upgrade failed: %v", err)
			}
		}
	}()

	addr := facades.Config().GetString("http.host") + ":" + facades.Config().GetString("http.port")
	ln, err := upg.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	go func() {
		if err = facades.Route().Listen(ln); !errors.Is(err, http.ErrServerClosed) {
			facades.Log().Errorf("HTTP server error: %v", err)
		}
	}()

	// tableflip ready
	if err = upg.Ready(); err != nil {
		return err
	}

	<-upg.Exit()

	// Make sure to set a deadline on exiting the process
	// after upg.Exit() is closed. No new upgrades can be
	// performed if the parent doesn't exit.
	time.AfterFunc(60*time.Second, func() {
		facades.Log().Error("Graceful shutdown timeout, force exit")
		os.Exit(1)
	})

	// Wait for connections to drain.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return facades.Route().Shutdown(ctx)
}
