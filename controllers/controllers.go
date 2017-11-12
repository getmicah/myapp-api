package controllers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/getmicah/myapp-api/config"
	"github.com/gorilla/securecookie"
)

// AppConfig : global app config
var AppConfig = config.Get("./config/config.json")

var (
	// ClientID : API identifier
	ClientID = os.Getenv(AppConfig.Env.SpotifyID)
	// ClientSecret : API secret key
	ClientSecret = os.Getenv(AppConfig.Env.SpotifySecret)
	// SecureStateCookie : hashed state cookie
	SecureStateCookie = GenerateSecureCookie()
	// SecureAccessTokenCookie : hashed access token cookie
	SecureAccessTokenCookie = GenerateSecureCookie()
	// SecureRefreshTokenCookie : hashed refresh token cookie
	SecureRefreshTokenCookie = GenerateSecureCookie()
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

// GenerateRandomString : create a random string with n length
func GenerateRandomString(n int) string {
	b := GenerateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b)
}

// GenerateRandomBytes : create a random []byte with n length
func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

// GenerateSecureCookie : generate a secure cookie
func GenerateSecureCookie() *securecookie.SecureCookie {
	hash := GenerateRandomBytes(16)
	block := GenerateRandomBytes(16)
	return securecookie.New(hash, block)
}

// CreateEncryptedCookie : create an encrypted cookie
func CreateEncryptedCookie(s *securecookie.SecureCookie, name string, value string, expiry time.Time) (*http.Cookie, error) {
	encoded, err := s.Encode(name, value)
	if err != nil {
		return nil, err
	}
	cookie := &http.Cookie{
		Name:     name,
		Value:    encoded,
		Path:     "/",
		Expires:  expiry,
		Secure:   AppConfig.Production,
		HttpOnly: true,
	}
	return cookie, nil
}

// ReadEncryptedCookie : read encrypted cookie
func ReadEncryptedCookie(r *http.Request, s *securecookie.SecureCookie, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	dur := time.Since(cookie.Expires).Seconds()
	if dur < 0 {
		return "", errors.New("timed_out")
	}
	var dst string
	decodeErr := s.Decode(name, cookie.Value, &dst)
	if decodeErr != nil {
		return "", decodeErr
	}
	return dst, nil
}

// ClearEncryptedCookie : set cookie(s) value empty
func ClearEncryptedCookie(w http.ResponseWriter, s *securecookie.SecureCookie, name string) error {
	expiry := time.Now().Add(-100 * time.Hour)
	cookie, err := CreateEncryptedCookie(s, name, "", expiry)
	if err != nil {
		return err
	}
	http.SetCookie(w, cookie)
	return nil
}

// RequestNewAccessToken : ask Spotify for new access_token with refresh token
func RequestNewAccessToken(r *http.Request) (*Token, error) {
	refreshToken, _ := ReadEncryptedCookie(r, SecureRefreshTokenCookie, AppConfig.Cookie.RefreshToken)
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", ClientID)
	data.Set("client_secret", ClientSecret)
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

// LoadAccessToken : load access_token cookie
func LoadAccessToken(w http.ResponseWriter, r *http.Request) (string, error) {
	accessToken, err := ReadEncryptedCookie(r, SecureAccessTokenCookie, AppConfig.Cookie.AccessToken)
	if err != nil {
		if err.Error() == "timed_out" {
			token, err := RequestNewAccessToken(r)
			if err != nil {
				return "", err
			}
			tokenExpiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
			cookie, err := CreateEncryptedCookie(SecureAccessTokenCookie, AppConfig.Cookie.AccessToken, token.AccessToken, tokenExpiry)
			http.SetCookie(w, cookie)
			return token.AccessToken, nil
		}
		return "", err
	}
	return accessToken, nil
}

// SendError : send ErrorResponse to user
func SendError(w http.ResponseWriter, code int, message string) {
	var e ErrorResponse
	e.Code = code
	e.Message = message
	body, _ := json.Marshal(e)
	w.WriteHeader(e.Code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// Get : GET API endpoint
func Get(w http.ResponseWriter, r *http.Request, endpoint string) {
	accessToken, err := LoadAccessToken(w, r)
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
