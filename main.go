package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	/**
			This is one of my least favorite things working with SQL in Go currently.
	You have to import the driver, but you don't use it directly anywhere in your
	code. The underscore tells Go that you're importing it for its side effects,
	not because you need to use it.
	*/
	_ "github.com/lib/pq"
	"github.com/rebyul/chirpy/internal/auth"
	"github.com/rebyul/chirpy/internal/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		errmsg := fmt.Errorf("failed to load env vars: %w", err)
		panic(errmsg)
	}

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	tokenSecret := os.Getenv("TOKEN_SECRET")

	db, dbErr := sql.Open("postgres", dbURL)

	if dbErr != nil {
		panic(dbErr)
	}

	dbQueries := database.New(db)
	serveMux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		platform:       platform,
		queries:        dbQueries,
		tokensecret:    tokenSecret,
	}

	fileHandler := fileHandler{}
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileHandler))
	serveMux.Handle("GET /api/healthz", healthHandler{})

	authHandlers := auth.AuthHandlers{
		Queries:     apiCfg.queries,
		TokenSecret: apiCfg.tokensecret,
	}
	serveMux.HandleFunc("POST /api/login", authHandlers.HandleLogin)
	serveMux.HandleFunc("POST /api/refresh", authHandlers.HandleRefresh)
	serveMux.HandleFunc("POST /api/revoke", authHandlers.HandleRevoke)

	metricHandler := metricHandler{cfg: &apiCfg}
	serveMux.Handle("GET /admin/metrics/", &metricHandler)

	chirpHandlers := ChirpHandlers{&apiCfg}
	serveMux.HandleFunc("GET /api/chirps", chirpHandlers.GetAllChirps)
	jwtMiddleware := auth.JwtAuthenticationMiddleware{
		Tokensecret: tokenSecret,
	}
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", chirpHandlers.GetChirpById)
	serveMux.Handle("POST /api/chirps", jwtMiddleware.MiddlewareJwtAuth(http.HandlerFunc(chirpHandlers.CreateChirp)))
	serveMux.Handle("DELETE /api/chirps/{chirpID}", jwtMiddleware.MiddlewareJwtAuth(http.HandlerFunc(chirpHandlers.DeleteChirpById)))

	userHandler := userHandler{cfg: &apiCfg}
	serveMux.HandleFunc("POST /api/users", userHandler.createUser)
	serveMux.Handle("PUT /api/users", jwtMiddleware.MiddlewareJwtAuth(http.HandlerFunc(userHandler.UpdateUser)))

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
