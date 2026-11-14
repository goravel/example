package main

import (
	"goravel/bootstrap"
)

//go:generate go run . artisan modular:publish

func main() {
	app := bootstrap.Boot()

	app.Start()
}
