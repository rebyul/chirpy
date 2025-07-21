package auth

import (
	"encoding/json"
	"errors"
	"io"
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
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}

	expiresIn := clampExpiresInMaxOneHour(req.ExpiresInSeconds)
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
		IsChirpyRed:  row.IsChirpyRed,
	}

	responses.SendJsonResponse(w, http.StatusOK, res)
}

func (a *AuthHandlers) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	refToken, err := GetBearerToken(r.Header)
	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "no refresh token in headers", err)
		return
	}

	found, err := a.Queries.GetRefreshToken(r.Context(), refToken)

	if err != nil || found.ExpiresAt.Before(time.Now().UTC()) {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "no refresh token in db", err)
		return
	}
	if found.RevokedAt.Valid {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "token has been revoked. cant refresh", err)
		return
	}

	newToken, err := MakeJWT(found.UserID, a.TokenSecret, clampExpiresInMaxOneHour(0))

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to create new token", err)
		return
	}

	type postRefreshTokenResponse struct {
		Token string `json:"token"`
	}

	responses.SendJsonResponse(w, http.StatusOK, postRefreshTokenResponse{
		Token: newToken,
	})

}

func (a *AuthHandlers) HandleRevoke(w http.ResponseWriter, r *http.Request) {

	refToken, err := GetBearerToken(r.Header)
	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "no refresh token in headers", err)
		return
	}

	found, err := a.Queries.GetRefreshToken(r.Context(), refToken)

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "no refresh token in db", err)
		return
	}

	if _, err := a.Queries.RevokeRefreshToken(r.Context(), found.Token); err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to revoke token in db", err)
		return
	}

	responses.SendJsonResponse(w, http.StatusNoContent, nil)
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
	authHeader := headers.Get("Authorization")

	// strip bearer
	index := strings.Index(strings.ToLower(authHeader), bearerPrefix)
	// Check bearer is at the beginning of the token
	if index != 0 {
		return "", ErrMissingBearerToken
	}

	token := authHeader[index+len(bearerPrefix):]
	return token, nil
}

const apiKeyPrefix = "apikey "

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")

	// strip apiKeyPrefix
	index := strings.Index(strings.ToLower(authHeader), apiKeyPrefix)
	// Check apiKeyPrefix is at the beginning of the token
	if index != 0 {
		return "", ErrMissingBearerToken
	}

	token := authHeader[index+len(apiKeyPrefix):]
	return token, nil
}
