package main

import (
	"fmt"
	"log"
	"os"

	"github.com/danielmiessler/fabric/internal/tui"
)

func main() {
	app, err := tui.NewTViewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Start(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}