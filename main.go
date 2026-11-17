package main

import (
	"goravel/bootstrap"
)

//go:generate go run . artisan module:publish

func main() {
	app := bootstrap.Boot()

	app.Start()
}
