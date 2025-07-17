package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rebyul/chirpy/internal/database"
	"github.com/rebyul/chirpy/internal/responses"
)

type AuthHandlers struct {
	Queries     *database.Queries
	TokenSecret string
}

func (a *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type postLoginRequest struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	var req postLoginRequest

	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to read body", err)
		return
	}

	if err := json.Unmarshal(data, &req); err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to unmarshal data", err)
		return
	}

	row, err := a.Queries.GetUserByEmail(r.Context(), req.Email)

	if passErr := CheckPasswordHash(req.Password, row.HashedPassword); err != nil || passErr != nil {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	type postLoginResponse struct {
		Id           string    `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	expiresIn := clampExpiresInMaxOneHour(req.ExpiresInSeconds)
	log.Printf("Req expire: %v, calc expire: %v", req.ExpiresInSeconds, expiresIn)
	token, err := MakeJWT(row.ID, a.TokenSecret, expiresIn)
	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to create jwt", err)
	}

	refreshToken, err := MakeRefreshToken()

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to create refresh token", err)
	}

	savedRefToken, err := a.Queries.CreateRefreshToken(r.Context(),
		database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    row.ID,
			ExpiresAt: time.Now().UTC().Add(60 * time.Hour * 24),
		})

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to save refresh token", err)
	}

	res := postLoginResponse{
		Id:           row.ID.String(),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		Email:        row.Email,
		Token:        token,
		RefreshToken: savedRefToken.Token,
	}

	responses.SendJsonResponse(w, http.StatusOK, res)
}

func clampExpiresInMaxOneHour(input int) time.Duration {
	hourDuration := time.Hour
	inputDuration := time.Duration(input) * time.Second

	// Default to 1 hour
	if input == 0 {
		return hourDuration
	}

	if inputDuration < hourDuration {
		return inputDuration
	}

	return hourDuration
}

var (
	ErrMissingBearerToken = errors.New("missing bearer token")
)

const bearerPrefix = "bearer "

func GetBearerToken(headers http.Header) (string, error) {
	bearer := headers.Get("Authorization")

	// strip bearer
	index := strings.Index(strings.ToLower(bearer), bearerPrefix)
	// Check bearer is at the beginning of the token
	if index != 0 {
		return "", ErrMissingBearerToken
	}

	token := bearer[index+len(bearerPrefix):]
	return token, nil
}
