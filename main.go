package main

import (
	"net/http"

	"github.com/getmicah/myapp-api/config"
	"github.com/getmicah/myapp-api/controllers"
	"github.com/rs/cors"
)

// AppConfig : global app config
var AppConfig = config.Get("./config/config.json")

func main() {
	// router
	mux := http.NewServeMux()
	// route /auth
	mux.HandleFunc("/auth/login", controllers.Login)
	mux.HandleFunc("/auth/logout", controllers.Logout)
	mux.HandleFunc("/auth/callback", controllers.Callback)
	// route /me
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		controllers.Get(w, r, "/me")
	})
	// middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowCredentials: true,
	})
	app := c.Handler(mux)
	// Go!
	http.ListenAndServe(":3000", app)
}
