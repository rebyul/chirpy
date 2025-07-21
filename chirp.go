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
	"github.com/rebyul/chirpy/internal/auth"
	"github.com/rebyul/chirpy/internal/database"
	"github.com/rebyul/chirpy/internal/responses"
)

const (
	chirpTooLongText string = "Chirp is too long"
)

type ChirpHandlers struct {
	cfg *apiConfig
}

type ChirpResponse struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    string    `json:"user_id"`
}

var ChirpHandler ChirpHandlers = ChirpHandlers{cfg: nil}

func (c *ChirpHandlers) CreateChirp(w http.ResponseWriter, r *http.Request) {
	// This this duplicated through auth middleware but i need the user id here
	// There must be a way to nicely pass down the userid/token down from the middleware to the handlers
	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	userId, err := auth.ValidateJWT(tok, c.cfg.tokensecret)

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	type createChirpRequest struct {
		Body string `json:""`
	}

	var req = createChirpRequest{}
	decoder := json.NewDecoder(r.Body)

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&req); err != nil {
		log.Printf("failed to decode parameters: %s", err)
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to decode body", err)
		return
	}

	// Validation
	sanitized, valid := getSanitizedChirp(req.Body)

	if !valid {
		responses.SendJsonErrorResponse(w, http.StatusBadRequest, chirpTooLongText, nil)
		return
	}

	// Save chirp
	saved, err := c.cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{Body: sanitized, UserID: userId})

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to save user to db", err)
		return
	}

	res := ChirpResponse{
		Id:        saved.ID.String(),
		CreatedAt: saved.CreatedAt,
		UpdatedAt: saved.UpdatedAt,
		Body:      saved.Body,
		UserId:    saved.UserID.String(),
	}

	// Create response
	responses.SendJsonResponse(w, http.StatusCreated, res)
}

func (c *ChirpHandlers) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	var filter_author = uuid.NullUUID{Valid: false}

	if queryParams.Has("author_id") {
		authorString := queryParams.Get("author_id")
		log.Println("found author id", authorString)
		parsedGuid, err := uuid.Parse(authorString)
		if err != nil {
			responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to retrieve chirps", err)
			return
		}
		filter_author = uuid.NullUUID{
			UUID:  parsedGuid,
			Valid: true,
		}
	}

	chirps, err := c.cfg.queries.GetChirps(r.Context(), filter_author)

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to retrieve chirps", err)
		return
	}

	chirpResponses := make([]ChirpResponse, 0, len(chirps))

	for _, c := range chirps {
		chirpResponses = append(chirpResponses, ChirpResponse{Id: c.ID.String(),
			CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, Body: c.Body, UserId: c.UserID.String()})
	}

	responses.SendJsonResponse(w, http.StatusOK, chirpResponses)
}

func (c *ChirpHandlers) GetChirpById(w http.ResponseWriter, r *http.Request) {
	chirpParam := r.PathValue("chirpID")

	chirpUuid, err := uuid.Parse(chirpParam)
	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid chirp id: %s", chirpParam), err)
		return
	}

	chirp, err := c.cfg.queries.GetChirpById(r.Context(), chirpUuid)

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusNotFound, fmt.Sprintf("couldn't find chirp id: %s", chirpParam), err)
		return
	}

	responses.SendJsonResponse(w, http.StatusOK, ChirpResponse{Id: chirp.ID.String(),
		CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body,
		UserId: chirp.UserID.String()})
	return
}

func (c *ChirpHandlers) DeleteChirpById(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	userId, err := auth.ValidateJWT(tok, c.cfg.tokensecret)

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	chirpParam := r.PathValue("chirpID")
	chirpUuid, err := uuid.Parse(chirpParam)
	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid chirp id: %s", chirpParam), err)
		return
	}

	chirp, err := c.cfg.queries.GetChirpById(r.Context(), chirpUuid)

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusNotFound, fmt.Sprintf("couldn't find chirp id: %s", chirpParam), err)
		return
	}

	if chirp.UserID != userId {
		responses.SendJsonErrorResponse(w, http.StatusForbidden, fmt.Sprintf("user doesn't own chirp id:%s", chirpParam), err)
		return
	}

	if err := c.cfg.queries.DeleteChirpById(r.Context(), chirpUuid); err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to delete chirp", err)
		return
	}

	responses.SendJsonResponse(w, http.StatusNoContent, nil)
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
