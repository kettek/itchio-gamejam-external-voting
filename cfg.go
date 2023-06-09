package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Address        string
	DBRoot         string
	GameJam        string
	GameJamName    string
	GameJamImage   string
	GameJamID      int
	ClientID       string
	OAuthRedirect  string
	VotingEnabled  bool
	VotingFinished bool
}

func loadConfig() Config {
	var config Config
	f, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, &config); err != nil {
		panic(err)
	}

	return config
}
