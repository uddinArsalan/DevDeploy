package utils

import (
	"encoding/json"
	"net/http"
)

type APIResonse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	SuccessMsg string      `json:"message,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(payload)
}

func SUCCESS(w http.ResponseWriter, status int, successMsg string, data interface{}) {
	JSON(w, status, APIResonse{
		Success:    true,
		Data:       data,
		SuccessMsg: successMsg,
	})
}

func FAIL(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, APIResonse{
		Success: false,
		Error:   msg,
	})
}
