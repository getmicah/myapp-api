package config

import (
	"encoding/json"
	"io/ioutil"
)

// AppConfig : global app properties
type AppConfig struct {
	Auth struct {
		AuthorizeURL string   `json:"authorizeURL"`
		TokenURL     string   `json:"tokenURL"`
		Redirect     string   `json:"redirect"`
		Scope        []string `json:"scope"`
	} `json:"auth"`
	Cookie struct {
		SpotifyAccessToken  string `json:"spotifyAccessToken"`
		SpotifyAuthState    string `json:"spotifyAuthState"`
		SpotifyRefreshToken string `json:"spotifyRefreshToken"`
	} `json:"cookie"`
	Env struct {
		SpotifyID     string `json:"spotifyID"`
		SpotifySecret string `json:"spotifySecret"`
	}
	Path   string `json:"path"`
	Port   int    `json:"port"`
	AppURL string `json:"appURL"`
}

// Get : open json file and return config
func Get(path string) AppConfig {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var c AppConfig
	json.Unmarshal(file, &c)
	return c
}
