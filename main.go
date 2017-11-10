package main

import (
	"net/http"

	"github.com/getmicah/myapp-api/config"
	"github.com/getmicah/myapp-api/controllers"
)

// AppConfig : global app config
var AppConfig = config.Get("./config/config.json")

func main() {
	http.HandleFunc("/authenticate", controllers.Authenticate)
	http.HandleFunc("/login", controllers.Login)
	http.HandleFunc("/logout", controllers.Logout)
	http.HandleFunc("/callback", controllers.Callback)
	http.ListenAndServe(":3000", nil)
}
