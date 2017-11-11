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
	mux := http.NewServeMux()
	mux.HandleFunc("/login", controllers.Login)
	mux.HandleFunc("/callback", controllers.Callback)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowCredentials: true,
	})
	handler := c.Handler(mux)
	http.ListenAndServe(":3000", handler)
}
