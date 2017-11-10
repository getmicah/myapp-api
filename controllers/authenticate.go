package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type authResponse struct {
	AccessToken *string `json:"accessToken"`
	Error       *string `json:"error"`
}

func sendAuthResponse(w http.ResponseWriter, accessToken string) {
	var ar authResponse
	ar.AccessToken = &accessToken
	ar.Error = nil
	json.NewEncoder(w).Encode(ar)
}

func sendAuthError(w http.ResponseWriter, msg string) {
	var ar authResponse
	ar.AccessToken = nil
	ar.Error = &msg
	json.NewEncoder(w).Encode(ar)
}

// Authenticate : check vailidity of access token
func Authenticate(w http.ResponseWriter, r *http.Request) {
	accessTokenCookie, err := r.Cookie(accessTokenCookieName)
	if err != nil {
		ClearCookies(w, authStateCookieName, accessTokenCookieName, refreshTokenCookieName)
		sendAuthError(w, "no access token")
		return
	}
	refreshTokenCookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		ClearCookies(w, authStateCookieName, accessTokenCookieName, refreshTokenCookieName)
		sendAuthError(w, "no refresh token")
		return
	}
	if time.Now().Sub(accessTokenCookie.Expires) < 0 {
		data := url.Values{}
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", refreshTokenCookie.Value)
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
		u, _ := url.ParseRequestURI(tokenURL)
		res, err := http.Post(u.String(), "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
		if err != nil {
			sendAuthError(w, err.Error())
		}
		var tr token
		if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
			sendAuthError(w, err.Error())
		}
		sendAuthResponse(w, tr.AccessToken)
		return
	}
	sendAuthResponse(w, accessTokenCookie.Value)
}
