package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%v]: %s\n", time.Now(), r.URL.Path)
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

type metricHandler struct {
	cfg *apiConfig
}

func (m metricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	template, err := os.ReadFile("./" + r.URL.Path + "index.html")

	if err != nil {
		panic(err)
	}

	strTemplate := string(template)
	times := strconv.Itoa(int(m.cfg.fileserverHits.Load()))
	strTemplate = strings.Replace(strTemplate, "%d", times, 1)

	if _, err := fmt.Fprintf(w, "%s", strTemplate); err != nil {
		fmt.Println("Failed to write response")
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
