package tests

import (
	"io"

	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/testing"

	"goravel/app/facades"
	"goravel/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}

// CaptureArtisanOutput runs the given artisan command and returns its output
// along with any error. Callers are responsible for asserting the error.
func (r *TestCase) CaptureArtisanOutput(command string) (string, error) {
	var err error
	output := color.CaptureOutput(func(_ io.Writer) {
		err = facades.Artisan().Call(command)
	})

	return output, err
}
