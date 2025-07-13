package main

import (
	"sync/atomic"

	"github.com/rebyul/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
}
