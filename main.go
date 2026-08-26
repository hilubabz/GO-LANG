package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/hilubabz/GO-LANG/middleware"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	hits := cfg.fileserverHits.Load()
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, hits)))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits: 0"))
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type validateChirpBody struct {
		Body string `json:"body"`
	}

	validator := validateChirpBody{}

	err := json.NewDecoder(r.Body).Decode(&validator)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if len(validator.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	type validateChirpResponse struct {
		Valid bool `json:"valid"`
	}

	response := validateChirpResponse{
		Valid: true,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func main() {
	var apiMiddleware apiConfig

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.Handle(
		"/app/",
		apiMiddleware.middlewareMetricsInc(
			http.StripPrefix(
				"/app/",
				http.FileServer(http.Dir(".")),
			),
		),
	)

	mux.HandleFunc("GET /admin/metrics", apiMiddleware.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiMiddleware.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

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