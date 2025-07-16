package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := setupWhatsAppClient("whatsapp_session.db"); err != nil {
		log.Printf("Could not start WhatsApp client: %v", err)
	}

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/api/status-updates", statusUpdateHandler)
	http.HandleFunc("/history", historyHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		fmt.Printf("WhatsApp Status Monitor by 0xagil starting on http://localhost:%s\n", port)
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			fmt.Printf("Error starting server: %s\n", err)
		}
	}()
	time.Sleep(2 * time.Second)

	select {}
}


