package main

import (
	"log"
	"net/http"
	"path/filepath"

	casper "github.com/tcnksm/go-casper"
)

const (
	defaultPort = "3000"
)

func main() {
	// See README.md to generate crt and key for this testing.
	certFile, _ := filepath.Abs("crts/server.crt")
	keyFile, _ := filepath.Abs("crts/server.key")

	// Initialize casper.
	pusher := casper.New(1<<6, 10)

	// Handle root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] %s %s", r.Method, r.URL.String())

		assets := []string{
			"/static/example.jpg",
		}

		// Server push!
		if _, err := pusher.Push(w, r, assets, nil); err != nil {
			log.Fatalf("[ERROR] Failed to push assets %v: %s", assets, err)
		}

		// Check what is pushed.
		if pushed := pusher.Pushed(); len(pushed) != 0 {
			log.Printf("[INFO] Pushed!: %v", pushed)
		}

		w.Header().Add("Content-Type", "text/html")
		w.Write([]byte(`<img src="/static/example.jpg"/>`))
	})

	// Handle static assets.
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] %s %s", r.Method, r.URL.String())
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))).ServeHTTP(w, r)
	})

	// Start listening.
	log.Println("[INFO] Start listening on port", ":"+defaultPort)
	if err := http.ListenAndServeTLS(":"+defaultPort, certFile, keyFile, nil); err != nil {
		log.Fatalf("[ERROR] Failed to listen: %s", err)
	}
}
