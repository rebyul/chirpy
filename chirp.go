package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rebyul/chirpy/internal/database"
)

type createChirpRequest struct {
	Body   string `json:"body"`
	UserId string `json:"user_id"`
}

type chirpCreateHandler struct {
	cfg *apiConfig
}

type chirpCreatedResponse struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    string    `json:"user_id"`
}

const (
	chirpTooLongText string = "Chirp is too long"
)

func (c *chirpCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req = createChirpRequest{}
	decoder := json.NewDecoder(r.Body)

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&req); err != nil {
		log.Printf("failed to decode parameters: %s", err)
		sendJsonErrorResponse(w, http.StatusInternalServerError, "failed to decode body", err)

		return
	}

	// Validation
	sanitized, valid := getSanitizedChirp(req.Body)

	if !valid {
		sendJsonErrorResponse(w, http.StatusBadRequest, chirpTooLongText, nil)
		return
	}

	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		sendJsonErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid userid guid: %s", req.UserId), err)
		return
	}

	// Save chirp
	saved, err := c.cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{Body: sanitized, UserID: userId})

	if err != nil {
		sendJsonErrorResponse(w, http.StatusInternalServerError, "failed to save user to db", err)
		return
	}

	res := chirpCreatedResponse{
		Id:        saved.ID.String(),
		CreatedAt: saved.CreatedAt,
		UpdatedAt: saved.UpdatedAt,
		Body:      saved.Body,
		UserId:    saved.UserID.String(),
	}

	// Create response
	sendJsonResponse(w, http.StatusCreated, res)
}

func getSanitizedChirp(chirp string) (string, bool) {
	if valid := validateChirp(chirp); !valid {
		return "", false
	}
	return replaceProfanities(chirp), true
}

func validateChirp(chirp string) bool {
	isValid := len(chirp) <= 140

	return isValid
}

func replaceProfanities(chirp string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	regexCombination := "(?i)(" + strings.Join(profaneWords, "|") + ")"
	pRegex := regexp.MustCompile(regexCombination)
	return pRegex.ReplaceAllString(chirp, "****")
}
