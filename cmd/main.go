package main

import "cakestore/internal/bootstrap"

func main() {
	app := bootstrap.NewApplication()
	app.Bootstrap()
	app.Start()
}
