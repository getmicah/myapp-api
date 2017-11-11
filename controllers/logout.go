package controllers

import "net/http"

// Logout : clear authentication cookies
func Logout(w http.ResponseWriter, r *http.Request) {
	ClearCookie(w, authStateCookieName)
	http.Redirect(w, r, appURL, 302)
}
