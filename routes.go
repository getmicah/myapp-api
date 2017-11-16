package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LoginHandler : /auth/login
type LoginHandler struct {
	clientID        string
	redirectURI     string
	scope           []string
	authStateCookie CookieID
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET /login")
		loginGet(w, r, h)
	default:
		SendBadRequest(w, r.Method)
	}
}

func loginGet(w http.ResponseWriter, r *http.Request, h *LoginHandler) {
	dur := 3600 * time.Second
	state := GenerateRandomString(16)
	expiry := time.Now().Add(dur)
	if err := WriteCookie(w, h.authStateCookie, state, expiry); err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
	}
	api := "https://accounts.spotify.com/authorize/"
	authURL := fmt.Sprintf(
		"%s?client_id=%s&response_type=%s&redirect_uri=%s&scope=%s&state=%s",
		api, h.clientID, "code", url.PathEscape(h.redirectURI), strings.Join(h.scope[:], "%20"), state,
	)
	http.Redirect(w, r, authURL, 302)
}

// CallbackHandler : /auth/callback
type CallbackHandler struct {
	authStateCookie    CookieID
	accessTokenCookie  CookieID
	refreshTokenCookie CookieID
	tokenExpiryCookie  CookieID
	clientID           string
	clientSecret       string
	redirectURI        string
	appURL             string
}

func (h *CallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET /callback")
		callbackGet(w, r, h)
	default:
		SendBadRequest(w, r.Method)
	}
}

func callbackGet(w http.ResponseWriter, r *http.Request, h *CallbackHandler) {
	originalState, err := ReadCookie(r, h.authStateCookie)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	newState := r.URL.Query().Get("state")
	if newState != originalState {
		SendError(w, http.StatusUnauthorized, "Auth state compormised")
		return
	}
	callbackErr := r.URL.Query().Get("error")
	if callbackErr != "" {
		SendError(w, http.StatusUnauthorized, callbackErr)
		return
	}
	code := r.URL.Query().Get("code")
	token, err := RequestOAuthToken(r, code, h.redirectURI, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	accessTokenExpiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	yearExpiry := time.Now().Add(365 * 24 * time.Hour)
	WriteCookie(w, h.accessTokenCookie, token.AccessToken, accessTokenExpiry)
	WriteCookie(w, h.refreshTokenCookie, token.RefreshToken, yearExpiry)
	WriteCookie(w, h.tokenExpiryCookie, accessTokenExpiry.Format(TimeLayout), yearExpiry)
	ClearCookie(w, h.authStateCookie)
	http.Redirect(w, r, h.appURL, 302)
}

// LogoutHandler : /auth/logout
type LogoutHandler struct {
	cookies []CookieID
	appURL  string
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET /logout")
		logoutGet(w, r, h)
	default:
		SendBadRequest(w, r.Method)
	}
}

func logoutGet(w http.ResponseWriter, r *http.Request, h *LogoutHandler) {
	for i := 0; i < len(h.cookies); i++ {
		if err := ClearCookie(w, h.cookies[i]); err != nil {
			SendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	http.Redirect(w, r, h.appURL, 302)
}

// AuthHandler : /auth
type AuthHandler struct {
	accessTokenCookie  CookieID
	refreshTokenCookie CookieID
	tokenExpiryCookie  CookieID
	clientID           string
	clientSecret       string
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET /auth")
		authGet(w, r, h)
	default:
		SendBadRequest(w, r.Method)
	}
}

func authGet(w http.ResponseWriter, r *http.Request, h *AuthHandler) {
	_, err := LoadAccessToken(w, r, h.accessTokenCookie, h.refreshTokenCookie, h.tokenExpiryCookie, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// SearchHandler : /search
type SearchHandler struct {
	accessTokenCookie  CookieID
	refreshTokenCookie CookieID
	tokenExpiryCookie  CookieID
	clientID           string
	clientSecret       string
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET /search")
		searchGet(w, r, h)
	default:
		SendBadRequest(w, r.Method)
	}

}

func searchGet(w http.ResponseWriter, r *http.Request, h *SearchHandler) {
	accessToken, err := LoadAccessToken(w, r, h.accessTokenCookie, h.refreshTokenCookie, h.tokenExpiryCookie, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	q := r.URL.Query().Get("q")
	searchType := r.URL.Query().Get("type")
	limit := 5
	api := fmt.Sprintf("/search?q=%s&type=%s&limit=%d&market=US", url.PathEscape(q), searchType, limit)
	res, err := SpotifyGet(r, api, accessToken)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// ArtistHandler : /artist
type ArtistHandler struct {
	accessTokenCookie  CookieID
	refreshTokenCookie CookieID
	tokenExpiryCookie  CookieID
	clientID           string
	clientSecret       string
}

func (h *ArtistHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET /artist")
		artistGet(w, r, h)
	default:
		SendBadRequest(w, r.Method)
	}

}

func artistGet(w http.ResponseWriter, r *http.Request, h *ArtistHandler) {
	accessToken, err := LoadAccessToken(w, r, h.accessTokenCookie, h.refreshTokenCookie, h.tokenExpiryCookie, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	id := r.URL.Query().Get("id")
	endpoint := fmt.Sprintf("/artists/%s", id)
	res, err := SpotifyGet(r, endpoint, accessToken)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// TrackHandler : /track
type TrackHandler struct {
	accessTokenCookie  CookieID
	refreshTokenCookie CookieID
	tokenExpiryCookie  CookieID
	clientID           string
	clientSecret       string
}

func (h *TrackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET /track")
		trackGet(w, r, h)
	default:
		SendBadRequest(w, r.Method)
	}

}

func trackGet(w http.ResponseWriter, r *http.Request, h *TrackHandler) {
	accessToken, err := LoadAccessToken(w, r, h.accessTokenCookie, h.refreshTokenCookie, h.tokenExpiryCookie, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	id := r.URL.Query().Get("id")
	endpoint := fmt.Sprintf("/tracks/%s", id)
	res, err := SpotifyGet(r, endpoint, accessToken)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// RecHandler : /rec
type RecHandler struct {
	accessTokenCookie  CookieID
	refreshTokenCookie CookieID
	tokenExpiryCookie  CookieID
	clientID           string
	clientSecret       string
}

func (h *RecHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("GET /rec")
		recGet(w, r, h)
	default:
		SendBadRequest(w, r.Method)
	}

}

func recGet(w http.ResponseWriter, r *http.Request, h *RecHandler) {
	accessToken, err := LoadAccessToken(w, r, h.accessTokenCookie, h.refreshTokenCookie, h.tokenExpiryCookie, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	url := fmt.Sprintf("/recommendations?market=US&limit=30&%s", r.URL.RawQuery)
	res, err := SpotifyGet(r, url, accessToken)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// PlaylistHandler : /rec
type PlaylistHandler struct {
	accessTokenCookie  CookieID
	refreshTokenCookie CookieID
	tokenExpiryCookie  CookieID
	clientID           string
	clientSecret       string
}

func (h *PlaylistHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		fmt.Println("POST /playlist")
		playlistPost(w, r, h)
	default:
		SendBadRequest(w, r.Method)
	}

}

func playlistPost(w http.ResponseWriter, r *http.Request, h *PlaylistHandler) {
	// get user access token
	accessToken, err := LoadAccessToken(w, r, h.accessTokenCookie, h.refreshTokenCookie, h.tokenExpiryCookie, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// get user id
	meRes, err := SpotifyGet(r, "/me", accessToken)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	var me User
	json.NewDecoder(meRes.Body).Decode(&me)

	// create user playlist
	userPlaylistEndpoint := fmt.Sprintf("/users/%s/playlists", me.ID)
	var pb PlaylistBody
	pb.Name = "RecPlaylist"
	playlistReqBody := new(bytes.Buffer)
	json.NewEncoder(playlistReqBody).Encode(pb)
	playlistReq, err := SpotifyPost(r, userPlaylistEndpoint, playlistReqBody, accessToken)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// add tracks from body to playlist
	var playlistResponse PlaylistResponse
	json.NewDecoder(playlistReq.Body).Decode(&playlistResponse)
	ptEndpoint := fmt.Sprintf("/users/%s/playlists/%s/tracks", me.ID, playlistResponse.ID)
	ptPostRes, err := SpotifyPost(r, ptEndpoint, r.Body, accessToken)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// return playlist id from spotify
	body, err := ioutil.ReadAll(ptPostRes.Body)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
