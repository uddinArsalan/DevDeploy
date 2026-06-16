package handlers

import (
	"encoding/json"
	"fmt"
	// "io"
	"net/http"
)

type UserDetails struct{
	Name  string `json:"name"`
	
}

func CreateAccount(w http.ResponseWriter,r *http.Request){
	var userDetails UserDetails
	err := json.NewDecoder(r.Body).Decode(&userDetails)
	if err != nil{
		http.Error(w,"Invalid details",http.StatusBadRequest)
		return
	}
	fmt.Printf("Creating account for user with name %v\n", userDetails)
}