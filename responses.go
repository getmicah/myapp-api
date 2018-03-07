package main

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

// AuthStatus : spotify authentication status
type AuthStatus struct {
	Authenticated bool `json:"authenticated"`
}

// User : /me spotify response
type User struct {
	ID string `json:"id"`
}

// PlaylistBody : post body for spotify playlist
type PlaylistBody struct {
	Name string `json:"name"`
}

// PlaylistResponse : spotify playlist
type PlaylistResponse struct {
	ID string `json:"id"`
}

// PlaylistTracksBody : post tracks to playlist
type PlaylistTracksBody struct {
	URIS []string `json:"uris"`
}

// PlaylistReturnJSON : return data for frontend
type PlaylistReturnJSON struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
