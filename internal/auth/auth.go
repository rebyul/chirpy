package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/rebyul/chirpy/internal/database"
	"github.com/rebyul/chirpy/internal/responses"
)

type AuthHandlers struct {
	Queries *database.Queries
}

func (a *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	type postLoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
		Id        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	res := postLoginResponse{
		Id:        row.Email,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		Email:     row.Email,
	}

	responses.SendJsonResponse(w, http.StatusOK, res)
}
