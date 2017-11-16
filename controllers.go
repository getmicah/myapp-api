package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// SendError : send an error response back the the user
func SendError(w http.ResponseWriter, code int, message string) {
	var e ErrorResponse
	e.Code = code
	e.Message = message
	body, err := json.Marshal(e)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(e.Code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// SendBadRequest : send a method error
func SendBadRequest(w http.ResponseWriter, method string) {
	msg := fmt.Sprintf("Endpoint doesn't support %s request", method)
	SendError(w, http.StatusBadRequest, msg)
}

// SpotifyGet : make a GET request to Spotify API
func SpotifyGet(r *http.Request, endpoint string, accessToken string) (*http.Response, error) {
	client := &http.Client{Timeout: ClientTimeout}
	u := fmt.Sprintf("https://api.spotify.com/v1%s", endpoint)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.New("Invalid access_token")
	}
	bearer := fmt.Sprintf("Bearer %s", accessToken)
	req.Header.Set("Authorization", bearer)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}
	return res, nil
}

// SpotifyPost : make a POST request to Spotify API
func SpotifyPost(r *http.Request, endpoint string, body io.Reader, accessToken string) (*http.Response, error) {
	client := &http.Client{Timeout: ClientTimeout}
	u := fmt.Sprintf("https://api.spotify.com/v1%s", endpoint)
	req, err := http.NewRequest("POST", u, body)
	if err != nil {
		return nil, err
	}
	bearer := fmt.Sprintf("Bearer %s", accessToken)
	req.Header.Set("Authorization", bearer)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if !(res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated) {
		return nil, errors.New(res.Status)
	}
	return res, nil
}

// SpotifyAuthPost : make a POST request to Spotify accounts API and receive a token
func SpotifyAuthPost(r *http.Request, body url.Values, clientID string, clientSecret string) (*Token, error) {
	client := &http.Client{Timeout: ClientTimeout}
	u := "https://accounts.spotify.com/api/token"
	req, err := http.NewRequest("POST", u, bytes.NewBufferString(body.Encode()))
	if err != nil {
		return nil, err
	}
	bearer := fmt.Sprintf("%s:%s", clientID, clientSecret)
	secret := base64.StdEncoding.EncodeToString([]byte(bearer))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", secret))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}
	var tr Token
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

// RequestOAuthToken : ask Spotify for an oauth token
func RequestOAuthToken(r *http.Request, code string, redirectURI string, clientID string, clientSecret string) (*Token, error) {
	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)
	token, err := SpotifyAuthPost(r, body, clientID, clientSecret)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// RequestNewOAuthToken : ask Spotify for a new oauth token
func RequestNewOAuthToken(r *http.Request, refreshToken string, clientID string, clientSecret string) (*Token, error) {
	body := url.Values{}
	body.Set("grant_type", "refresh_token")
	body.Set("refresh_token", refreshToken)
	token, err := SpotifyAuthPost(r, body, clientID, clientSecret)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// LoadAccessToken : load acces token from cookies
func LoadAccessToken(w http.ResponseWriter, r *http.Request, accessTokenCookie CookieID, refreshTokenCookie CookieID, tokenExpiryCookie CookieID, clientID string, clientSecret string) (string, error) {
	tokenExpiry, err := ReadCookie(r, tokenExpiryCookie)
	if err != nil {
		return "", err
	}
	expiryTime, err := time.Parse(TimeLayout, tokenExpiry)
	if err != nil {
		return "", err
	}
	if time.Since(expiryTime) > 0 {
		refreshToken, err := ReadCookie(r, refreshTokenCookie)
		if err != nil {
			return "", err
		}
		token, err := RequestNewOAuthToken(r, refreshToken, clientID, clientSecret)
		if err != nil {
			return "", err
		}
		newTokenExpiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
		if err := WriteCookie(w, accessTokenCookie, token.AccessToken, newTokenExpiry); err != nil {
			return "", err
		}
		tokenExpiryValue := newTokenExpiry.Format(TimeLayout)
		yearExpiry := time.Now().Add(365 * 24 * time.Hour)
		if err := WriteCookie(w, tokenExpiryCookie, tokenExpiryValue, yearExpiry); err != nil {
			return "", err
		}
		return token.AccessToken, nil
	}
	accessToken, err := ReadCookie(r, accessTokenCookie)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}
