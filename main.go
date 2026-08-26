package main

import (
	"log"
	"net/http"
	"time"

	"github.com/hilubabz/GO-LANG/middleware"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))

	s := &http.Server{
		Addr:           ":8080",
		Handler:        middleware.MiddlewareLog(mux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Server starting on http://localhost:8080")

	log.Fatal(s.ListenAndServe())
}