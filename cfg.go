package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type Config struct {
	original       *Config
	Address        string
	DBRoot         string `json:",omitempty"`
	GameJam        string
	GameJamName    string `json:",omitempty"`
	GameJamImage   string `json:",omitempty"`
	GameJamID      int    `json:",omitempty"`
	ClientID       string
	OAuthRedirect  string
	VotingEnabled  bool
	VotingFinished bool
	Admins         []UserInfo `json:",omitempty"`
}

func loadConfig() Config {
	var config Config
	f, err := os.Open("config.json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Create a default.
			b, err := json.MarshalIndent(Config{
				Address: ":3000",
			}, "", "	")
			if err != nil {
				panic(err)
			}
			if err := ioutil.WriteFile("config.json", b, 644); err != nil {
				panic(err)
			}
			f, err = os.Open("config.json")
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, &config); err != nil {
		panic(err)
	}

	c2 := config
	config.original = &c2

	return config
}

func setConfig(key string, value string) {
	switch key {
	case "ClientID":
		c.ClientID = value
		c.original.ClientID = value
	case "OAuthRedirect":
		c.OAuthRedirect = value
		c.original.OAuthRedirect = value
	case "GameJam":
		c.GameJam = value
		c.original.GameJam = value
		// Hmm, there might be thread issues if a client HTTP connection is submitting to the DB exactly when this is called.
		db.Close()
		db = loadDB()
		loadJamInfo()
		entries = getEntries(c.GameJamID)
	case "VotingFinished":
		if value == "true" {
			c.VotingFinished = true
		} else {
			c.VotingFinished = false
		}
		c.original.VotingFinished = c.VotingFinished
	case "VotingEnabled":
		if value == "true" {
			c.VotingEnabled = true
		} else {
			c.VotingEnabled = false
		}
		c.original.VotingEnabled = c.VotingEnabled
	}
}

func saveConfig() error {
	b, err := json.MarshalIndent(c.original, "", "	")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("config.json", b, 644)
}
