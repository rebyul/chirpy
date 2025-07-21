package main

import (
	"sync/atomic"

	"github.com/rebyul/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
	platform       string
	tokensecret    string
	polkakey       string
}
