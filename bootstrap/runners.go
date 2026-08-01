package bootstrap

import (
	"fmt"

	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"

	"goravel/app/facades"
)

// QueueWorkerRunner runs a single queue worker with a fixed configuration.
// Unlike the framework's default queue runner (which consumes the
// queue.default connection), this runner targets an explicit connection and
// queue, allowing the example app to run several workers concurrently.
type QueueWorkerRunner struct {
	signature string
	args      contractsqueue.Args
	worker    contractsqueue.Worker
}

// NewQueueWorkerRunner creates a runner for the given connection/queue pair.
// The worker is created lazily in Run, so facades need not be ready at
// construction time (runners may be registered during package init).
func NewQueueWorkerRunner(signature string, args contractsqueue.Args) *QueueWorkerRunner {
	return &QueueWorkerRunner{
		signature: signature,
		args:      args,
	}
}

// Signature uniquely identifies the runner among auto-run runners and the
// disabled_runners config patterns.
func (r *QueueWorkerRunner) Signature() string {
	return r.signature
}

// ShouldRun reports whether the target connection is backed by a real driver
// (not sync). Sync connections process jobs in-process, so a worker would
// just idle.
func (r *QueueWorkerRunner) ShouldRun() bool {
	connection := r.args.Connection
	if connection == "" {
		connection = facades.Config().GetString("queue.default")
	}
	if connection == "" {
		return false
	}

	return facades.Config().GetString(fmt.Sprintf("queue.connections.%s.driver", connection)) != "sync"
}

// Run starts the worker, blocking until Shutdown is called.
func (r *QueueWorkerRunner) Run() error {
	r.worker = facades.Queue().Worker(r.args)

	return r.worker.Run()
}

// Shutdown stops the worker started by Run.
func (r *QueueWorkerRunner) Shutdown() error {
	if r.worker == nil {
		return nil
	}

	return r.worker.Shutdown()
}

// DefaultRunners returns the app-defined runners started on boot. The queue
// workers below were previously launched inline in the WithCallback block of
// Boot; moving them into runners lets the framework manage their lifecycle
// (start/stop on app restart) and honors the disabled_runners config.
func DefaultRunners() []contractsfoundation.Runner {
	return []contractsfoundation.Runner{
		NewQueueWorkerRunner("app:queue:database", contractsqueue.Args{
			Connection: "database",
		}),
		NewQueueWorkerRunner("app:queue:test", contractsqueue.Args{
			Connection: "database",
			Queue:      "test",
		}),
	}
}
