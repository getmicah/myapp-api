package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func getAuthToken(r *http.Request) (*Token, error) {
	authErr := r.URL.Query().Get("error")
	if authErr != "" {
		return nil, errors.New(authErr)
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, errors.New("invalid code")
	}
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", AppConfig.Auth.Redirect)
	data.Set("client_id", ClientID)
	data.Set("client_secret", ClientSecret)
	u := "https://accounts.spotify.com/api/token"
	res, err := http.Post(u, "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var tr Token
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

// Callback : save access token and redirect to index
func Callback(w http.ResponseWriter, r *http.Request) {
	originalState, err := ReadEncryptedCookie(r, SecureStateCookie, AppConfig.Cookie.State)
	if err != nil {
		// error header
		fmt.Println(err)
		return
	}
	newState := r.URL.Query().Get("state")
	if newState != originalState {
		// error header
		fmt.Println("new state != original state")
		return
	}
	token, err := getAuthToken(r)
	if err != nil {
		// error header
		fmt.Println(err)
		return
	}
	ClearEncryptedCookie(w, SecureStateCookie, AppConfig.Cookie.State)
	tokenExpiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	yearExpiry := time.Now().Add(365 * 24 * time.Hour)
	accessCookie, _ := CreateEncryptedCookie(SecureAccessTokenCookie, AppConfig.Cookie.AccessToken, token.AccessToken, tokenExpiry)
	refreshCookie, _ := CreateEncryptedCookie(SecureRefreshTokenCookie, AppConfig.Cookie.RefreshToken, token.RefreshToken, yearExpiry)
	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
	http.Redirect(w, r, AppConfig.AppURL, 302)
}
