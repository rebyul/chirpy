package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type userHandler struct {
	cfg *apiConfig
}

func (u *userHandler) createUser(w http.ResponseWriter, r *http.Request) {
	type postUserRequest struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var request postUserRequest
	if err := decoder.Decode(&request); err != nil {
		sendJsonErrorResponse(w, http.StatusInternalServerError, "failed to decode post body req", err)
		return
	}

	user, err := u.cfg.queries.CreateUser(r.Context(), request.Email)
	if err != nil {
		sendJsonErrorResponse(w, http.StatusInternalServerError, "db failed to save user", err)
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

	sendJsonResponse(w, http.StatusCreated, resp)
}
