package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

const (
	internalServerText string = "Something went wrong"
)

var (
	internalServerResponse = errorResponse{Error: internalServerText}
)

func sendJsonErrorResponse(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}

	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
		sendJsonResponse(w, code, internalServerResponse)
		return
	}

	sendJsonResponse(w, code, errorResponse{
		Error: msg,
	})
}

func sendJsonResponse(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")

	res, err := json.Marshal(payload)

	if err != nil {
		log.Printf("error marshalling json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(res)
}
