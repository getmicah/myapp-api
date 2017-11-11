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

type token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func getAuthToken(r *http.Request, originalState string) (*token, error) {
	newState := r.URL.Query().Get("state")
	if newState != originalState {
		return nil, errors.New("state error")
	}
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
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	u, _ := url.ParseRequestURI(tokenURL)
	res, err := http.Post(u.String(), "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	var tr token
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

// Callback : save access token and redirect to index
func Callback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie(authStateCookieName)
	if err != nil {
		// error header
		fmt.Println(err)
		return
	}
	ClearCookie(w, authStateCookieName)
	token, err := getAuthToken(r, state.Value)
	if err != nil {
		// error header
		fmt.Println(err)
		return
	}
	accessTokenExpiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Format(time.RFC1123)
	instantExpiry := time.Now().Add(1 * time.Second)
	accessTokenCookie := CreateCookie(accessTokenCookieName, token.AccessToken, instantExpiry, false)
	refreshTokenCookie := CreateCookie(refreshTokenCookieName, token.RefreshToken, instantExpiry, false)
	tokenExpiryCookie := CreateCookie(tokenExpiryCookieName, accessTokenExpiry, instantExpiry, false)
	http.SetCookie(w, accessTokenCookie)
	http.SetCookie(w, refreshTokenCookie)
	http.SetCookie(w, tokenExpiryCookie)
	http.Redirect(w, r, appURL, 302)
}
