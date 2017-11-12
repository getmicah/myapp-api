package controllers

import "net/http"
import "fmt"

// Logout : remove authentication cookies
func Logout(w http.ResponseWriter, r *http.Request) {
	if err := ClearEncryptedCookie(w, SecureAccessTokenCookie, AppConfig.Cookie.AccessToken); err != nil {
		// send error response
		fmt.Println(err)
		return
	}
	if err := ClearEncryptedCookie(w, SecureRefreshTokenCookie, AppConfig.Cookie.RefreshToken); err != nil {
		// send error response
		fmt.Println(err)
		return
	}
	http.Redirect(w, r, AppConfig.AppURL, 302)
}
