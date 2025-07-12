package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type chirpValidationHandler struct{}

type chirpBody struct {
	Body string `json:"body"`
}

type chirpValidationResponse struct {
	Valid bool `json:"valid"`
}

type errorResponse struct {
	Error string `json:"error"`
}

const (
	internalServerText string = "Something went wrong"
	chirpTooLongText   string = "Chirp is too long"
)

var (
	validChirpResponse     = chirpValidationResponse{Valid: true}
	chirpTooLongResponse   = errorResponse{Error: chirpTooLongText}
	internalServerResponse = errorResponse{Error: internalServerText}
)

func (c chirpValidationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var body = chirpBody{}
	decoder := json.NewDecoder(r.Body)

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&body); err != nil {
		log.Printf("failed to decode parameters: %s", err)
		res, err := json.Marshal(internalServerResponse)

		if err != nil {
			catastrophicResponse(w)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write(res); err != nil {
			catastrophicResponse(w)
			return
		}

		return
	}

	isValid := len(body.Body) <= 140
	err := sendChirpResponse(isValid, w)

	if err != nil {
		catastrophicResponse(w)
		return
	}
}

func sendChirpResponse(isValid bool, w http.ResponseWriter) error {
	switch isValid {
	case true:
		w.WriteHeader(http.StatusOK)
		res, err := json.Marshal(validChirpResponse)
		if err != nil {
			return err
		}

		w.Write(res)
		return nil
	case false:
		w.WriteHeader(http.StatusBadRequest)
		res, err := json.Marshal(chirpTooLongResponse)
		if err != nil {
			return err
		}

		w.Write(res)
		return nil
	}

	return nil

}
func catastrophicResponse(w http.ResponseWriter) {
	res, err := json.Marshal(internalServerResponse)

	if err != nil {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Catastrophic error: %v", err)
		panic(err)
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write(res)

}
