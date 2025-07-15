package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rebyul/chirpy/internal/auth"
	"github.com/rebyul/chirpy/internal/database"
	"github.com/rebyul/chirpy/internal/responses"
)

type userHandler struct {
	cfg *apiConfig
}

func (u *userHandler) createUser(w http.ResponseWriter, r *http.Request) {
	type postUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var request postUserRequest
	if err := decoder.Decode(&request); err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to decode post body req", err)
		return
	}

	hashedPw, err := auth.HashPassword(request.Password)

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to hash pw", err)
		return
	}

	user, err := u.cfg.queries.CreateUser(r.Context(),
		database.CreateUserParams{
			Email:          request.Email,
			HashedPassword: hashedPw,
		},
	)

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "db failed to save user", err)
		return
	}

	type postUserResponse struct {
		Id        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	resp := postUserResponse{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	responses.SendJsonResponse(w, http.StatusCreated, resp)
}

func isRawPwValid(pw string) bool {
	return len(pw) == 0
}
