package main

import (
	"log"

	"github.com/strawberry-code/mcp-curator/internal/ui"
)

func main() {
	app, err := ui.NewApp()
	if err != nil {
		log.Fatalf("Errore avvio applicazione: %v", err)
	}

	app.Run()
}
