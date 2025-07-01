package main

import (
	"os"
	"testing"

	"goravel/bootstrap"
)

func TestMain(m *testing.M) {
	bootstrap.Boot()
	os.Exit(m.Run())
}
