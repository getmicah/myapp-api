package main

import (
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

// CookieID : cookie identification
type CookieID struct {
	secure *securecookie.SecureCookie
	name   string
}

// GenerateCookie : generate a secure cookie
func GenerateCookie(name string) CookieID {
	var c CookieID
	hash := GenerateRandomBytes(16)
	block := GenerateRandomBytes(16)
	c.secure = securecookie.New(hash, block)
	c.name = name
	return c
}

// WriteCookie : create an http cookie
func WriteCookie(w http.ResponseWriter, c CookieID, value string, expiry time.Time) error {
	encoded, err := c.secure.Encode(c.name, value)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:    c.name,
		Value:   encoded,
		Path:    "/",
		Expires: expiry,
		//Secure:   true
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	return nil
}

// ReadCookie : read encrypted cookie
func ReadCookie(r *http.Request, c CookieID) (string, error) {
	httpCookie, err := r.Cookie(c.name)
	if err != nil {
		return "", err
	}
	var dst string
	decodeErr := c.secure.Decode(c.name, httpCookie.Value, &dst)
	if decodeErr != nil {
		return "", decodeErr
	}
	return dst, nil
}

// ClearCookie : set cookie(s) value empty
func ClearCookie(w http.ResponseWriter, c CookieID) error {
	expiry := time.Now().Add(-100 * time.Hour)
	if err := WriteCookie(w, c, "", expiry); err != nil {
		return err
	}
	return nil
}
