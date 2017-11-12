package controllers

import (
	"net/http"
)

// User : get /me endpoint
func User(w http.ResponseWriter, r *http.Request) {
	Get(w, r, "/me")
}
