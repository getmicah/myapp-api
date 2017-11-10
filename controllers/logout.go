package controllers

import "net/http"

// Logout : clear authentication cookies
func Logout(w http.ResponseWriter, r *http.Request) {
	ClearCookies(w, authStateCookieName, accessTokenCookieName, refreshTokenCookieName)
	http.Redirect(w, r, appURL, 302)
}
