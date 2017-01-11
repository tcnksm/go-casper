package main

import (
	"log"
	"net/http"
	"path/filepath"

	casper "github.com/tcnksm/go-casper"
)

func main() {
	certFile, _ := filepath.Abs("crts/server.crt")
	keyFile, _ := filepath.Abs("crts/server.key")

	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		csp := casper.New(1<<6, 10)
		if _, err := csp.Push(w, r, "/static/example.jpg", &casper.Options{}); err != nil {
			log.Fatal(err)
		}
		if _, err := csp.Push(w, r, "/static/example2.jpg", &casper.Options{}); err != nil {
			log.Fatal(err)
		}

		w.Header().Add("Content-Type", "text/html")
		w.Write([]byte(`<img src="/static/example.jpg"/>`))
	})

	log.Println(":3000")
	if err := http.ListenAndServeTLS(":3000", certFile, keyFile, nil); err != nil {
		log.Printf("[ERROR] %s", err)
	}
}
