package main

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

func respondError(w http.ResponseWriter, status int, err error) {
	body := ErrorResponse{
		ErrorMessage: err.Error(),
	}
	err = json.NewEncoder(w).Encode(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
	w.WriteHeader(status)
}
