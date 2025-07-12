package main

import (
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	serveMux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	apiCfg := apiConfig{fileserverHits: atomic.Int32{}}

	// fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	fileHandler := fileHandler{}
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileHandler))
	serveMux.Handle("GET /api/healthz", healthHandler{})
	metricHandler := metricHandler{&apiCfg}
	serveMux.Handle("GET /admin/metrics/", metricHandler)
	serveMux.Handle("POST /api/validate_chirp", chirpValidationHandler{})

	resetHandler := resetHandler{&apiCfg}
	serveMux.Handle("POST /admin/reset", resetHandler)

	err := server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}

type fileHandler struct{}

func (fileHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	http.StripPrefix("/app/", http.FileServer(http.Dir("."))).ServeHTTP(writer, request)
}
