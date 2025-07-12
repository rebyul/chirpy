package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type chirpValidationHandler struct{}

type chirpBody struct {
	Body string `json:"body"`
}

type chirpValidationResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

const (
	chirpTooLongText string = "Chirp is too long"
)

func (c chirpValidationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var body = chirpBody{}
	decoder := json.NewDecoder(r.Body)

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&body); err != nil {
		log.Printf("failed to decode parameters: %s", err)
		sendJsonErrorResponse(w, http.StatusInternalServerError, "failed to decode body", err)

		return
	}

	// Validation
	isValid := len(body.Body) <= 140

	switch isValid {
	case true:
		censoredBody := replaceProfanities(body.Body)
		sendJsonResponse(w, http.StatusOK, chirpValidationResponse{CleanedBody: censoredBody})
	case false:
		sendJsonErrorResponse(w, http.StatusBadRequest, chirpTooLongText, nil)
	}
}

func replaceProfanities(chirp string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	regexCombination := "(?i)(" + strings.Join(profaneWords, "|") + ")"
	pRegex := regexp.MustCompile(regexCombination)
	return pRegex.ReplaceAllString(chirp, "****")
}
