//go:build windows

package main

import (
	"log"

	"github.com/bodrovis-learning/kruk-shot/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
