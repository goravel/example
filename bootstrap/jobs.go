package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/jobs"
)

func Jobs() []queue.Job {
	return []queue.Job{
		&jobs.Test{},
		&jobs.TestErr{},
	}
}
