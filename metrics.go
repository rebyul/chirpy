package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("meow", r.RequestURI)
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

type metricHandler struct {
	cfg *apiConfig
}

func (m metricHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %d\n", m.cfg.fileserverHits.Load())
}

type resetHandler struct {
	cfg *apiConfig
}

func (r resetHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	r.cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0\n"))
}
