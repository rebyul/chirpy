package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
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
		Id          string    `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	resp := postUserResponse{
		Id:          user.ID.String(),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	responses.SendJsonResponse(w, http.StatusCreated, resp)
}

func (u *userHandler) UpdateUserEmailPassword(w http.ResponseWriter, r *http.Request) {
	type putUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusUnauthorized, "invalid or missing bearer token", err)
		return
	}
	userId, err := auth.ValidateJWT(token, u.cfg.tokensecret)
	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "invalid jwt token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var request putUserRequest
	if err := decoder.Decode(&request); err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to decode put body req", err)
		return
	}

	hashedPw, err := auth.HashPassword(request.Password)

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to hash pw", err)
		return
	}

	user, err := u.cfg.queries.UpdateUserEmailAndPassword(r.Context(),
		database.UpdateUserEmailAndPasswordParams{
			ID:             userId,
			Email:          request.Email,
			HashedPassword: hashedPw,
		},
	)

	if err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "db failed to update user", err)
		return
	}

	type postUserResponse struct {
		Id          string    `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	resp := postUserResponse{
		Id:          user.ID.String(),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	responses.SendJsonResponse(w, http.StatusOK, resp)
}

type PolkaRequestType int

const (
	UserUpgraded PolkaRequestType = iota
)

// func (p PolkaRequestType) String() string {
// 	return polkaRequestType[p]
// }
//
// var polkaRequestType = map[PolkaRequestType]string{
// 	UserUpgraded: "user.upgraded",
// }

var polkaReverseRequest = map[string]PolkaRequestType{
	"user.upgraded": UserUpgraded,
}

func (u *userHandler) UpgradeUserToChipyRed(w http.ResponseWriter, r *http.Request) {
	type polkaRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		}
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var req polkaRequest

	if err := decoder.Decode(&req); err != nil {
		responses.SendJsonErrorResponse(w, http.StatusInternalServerError, "failed to decode req", err)
	}

	switch polkaReverseRequest[req.Event] {
	case UserUpgraded:
		{

			log.Printf("decoded user id: %v\n", req.Data.UserId)
			userId, err := uuid.Parse(req.Data.UserId)
			if err != nil {
				responses.SendJsonErrorResponse(w, http.StatusNotFound, fmt.Sprintf("no user found with id: %s\n", req.Data.UserId), err)
				return
			}
			if _, err := u.cfg.queries.GetUserById(r.Context(), userId); err != nil {
				responses.SendJsonErrorResponse(w, http.StatusNotFound, fmt.Sprintf("no user found with id: %s\n", req.Data.UserId), err)
				return
			}

			if err := u.cfg.queries.UpgradeUserToChirpyRed(r.Context(), userId); err != nil {
				responses.SendJsonErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to update user: %s\n", req.Data.UserId), err)
				return
			}
			responses.SendJsonResponse(w, http.StatusNoContent, nil)
			return
		}
		// Return 204 if type is unknown
	default:
		{
			responses.SendJsonResponse(w, http.StatusNoContent, nil)
		}
	}

}
