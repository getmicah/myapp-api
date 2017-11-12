package main

import (
	"encoding/json"
	"io/ioutil"
)

// AppConfig : global app properties
type AppConfig struct {
	Auth struct {
		Redirect string   `json:"redirect"`
		Scope    []string `json:"scope"`
	} `json:"auth"`
	Env struct {
		SpotifyID     string `json:"spotifyID"`
		SpotifySecret string `json:"spotifySecret"`
	}
	Path       string `json:"path"`
	Port       int    `json:"port"`
	APIURL     string `json:"apiURL"`
	AppURL     string `json:"appURL"`
	Production bool   `json:"production"`
}

// GetConfig : open json file and return config
func GetConfig(path string) AppConfig {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var c AppConfig
	json.Unmarshal(file, &c)
	return c
}
