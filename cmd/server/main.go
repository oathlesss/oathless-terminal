package main

import (
	"log"
	"net/http"
	"os"

	"github.com/oathlesss/oathless-terminal/internal/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// API
	api := handler.NewAPI()
	mux.HandleFunc("GET /api/commands", api.HandleCommands)
	mux.HandleFunc("POST /api/command", api.HandleCommand)

	// SPA fallback (serves embedded Vue dist + index.html for all other routes)
	spa := handler.NewSPA()
	mux.HandleFunc("/", spa.Serve)

	log.Printf("oathless-terminal listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
