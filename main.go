// Package main is the entry point for the wave screensaver application.
package main

import (
	"log"

	"github.com/olegchuev/screensaver/internal/app"
)

// main initializes and runs the screensaver application.
func main() {
	cfg := app.DefaultConfig()
	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}

