package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
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

func GenerateRandomID() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func Slugify(name string) string {
	fields := strings.Fields(strings.ToLower(name))
	return strings.Join(fields, "-")
}
