package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Token : oauth2 token
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// ErrorResponse : http error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GetSpotify : make a GET request to Spotify API
func GetSpotify(w http.ResponseWriter, r *http.Request, endpoint string, accessTokenCookie CookieID) {
	accessToken, err := ReadCookie(r, accessTokenCookie)
	if err != nil {
		SendError(w, http.StatusUnauthorized, "Invalid access_token")
		return
	}
	client := &http.Client{}
	u := fmt.Sprintf("https://api.spotify.com/v1%s", endpoint)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Invalid access_token")
		return
	}
	bearer := fmt.Sprintf("Bearer %s", accessToken)
	req.Header.Set("Authorization", bearer)
	res, err := client.Do(req)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// SendError : send and error response back the the user
func SendError(w http.ResponseWriter, code int, message string) {
	var e ErrorResponse
	e.Code = code
	e.Message = message
	body, _ := json.Marshal(e)
	w.WriteHeader(e.Code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// RequestNewAccessToken : ask Spotify for a new access_token
func RequestNewAccessToken(r *http.Request, refreshToken string, clientID string, clientSecret string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	u := "https://accounts.spotify.com/api/token"
	res, err := http.Post(u, "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	var tr Token
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

// GetAuthToken : get Spotify oauth2 token
func GetAuthToken(r *http.Request, redirect string, clientID string, clientSecret string) (*Token, error) {
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
	data.Set("redirect_uri", redirect)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
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

// LoadAccessToken : load acces token from cookies
func LoadAccessToken(w http.ResponseWriter, r *http.Request, accessTokenCookie CookieID, tokenExpiryCookie CookieID, refreshToken string, clientID string, clientSecret string, timeLayout string) (string, error) {
	accessToken, err := ReadCookie(r, accessTokenCookie)
	if err != nil {
		if err.Error() == "timed_out" {
			token, err := RequestNewAccessToken(r, refreshToken, clientID, clientSecret)
			if err != nil {
				return "", err
			}
			tokenExpiryValue := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Format(timeLayout)
			yearExpiry := time.Now().Add(365 * 24 * time.Hour)
			if err := WriteCookie(w, tokenExpiryCookie, tokenExpiryValue, yearExpiry); err != nil {
				return "", err
			}
			return token.AccessToken, nil
		}
		return "", err
	}
	return accessToken, nil
}
