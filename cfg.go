package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
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
	BaseURL        string
	VoteCategories []string
	Badges         []string
	Tags           []string
	Admins         []UserInfo `json:",omitempty"`
	//
	Font     string
	FontSize float64
	//
	Badge struct {
		Width  int
		Height int
		TextX  int
		TextY  int
	}
	Tag struct {
		Width  int
		Height int
		TextX  int
		TextY  int
	}
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
			if err := os.WriteFile("config.json", b, 0644); err != nil {
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
	copy(config.original.VoteCategories, config.VoteCategories)
	copy(config.original.Admins, config.Admins)

	return config
}

func setConfig(key string, value string) error {
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
	case "AddVoteCategory":
		c.VoteCategories = append(c.VoteCategories, value)
		c.original.VoteCategories = append(c.original.VoteCategories, value)
	case "RemoveVoteCategory":
		if strings.HasPrefix(value, "VoteCategories-") {
			parts := strings.Split(value, "-")
			if len(parts) != 2 {
				return errors.New("bad vote index")
			}
			i, err := strconv.Atoi(parts[1])
			if err != nil {
				return errors.New("bad vote index")
			}
			if i < 0 || i >= len(c.VoteCategories) {
				return errors.New("bad vote index")
			}
			c.VoteCategories = append(c.VoteCategories[:i], c.VoteCategories[i+1:]...)
			c.original.VoteCategories = append(c.original.VoteCategories[:i], c.original.VoteCategories[i+1:]...)
		}
	case "AddTag":
		c.Tags = append(c.Tags, value)
		c.original.Tags = append(c.original.Tags, value)
	case "RemoveTag":
		if strings.HasPrefix(value, "Tags-") {
			parts := strings.Split(value, "-")
			if len(parts) != 2 {
				return errors.New("bad tag index")
			}
			i, err := strconv.Atoi(parts[1])
			if err != nil {
				return errors.New("bad tag index")
			}
			if i < 0 || i >= len(c.Tags) {
				return errors.New("bad tag index")
			}
			c.Tags = append(c.Tags[:i], c.Tags[i+1:]...)
			c.original.Tags = append(c.original.Tags[:i], c.original.Tags[i+1:]...)
		}
	case "AddBadge":
		c.Badges = append(c.Badges, value)
		c.original.Badges = append(c.original.Badges, value)
	case "RemoveBadge":
		if strings.HasPrefix(value, "Badges-") {
			parts := strings.Split(value, "-")
			if len(parts) != 2 {
				return errors.New("bad badge index")
			}
			i, err := strconv.Atoi(parts[1])
			if err != nil {
				return errors.New("bad badge index")
			}
			if i < 0 || i >= len(c.Badges) {
				return errors.New("bad badge index")
			}
			c.Badges = append(c.Badges[:i], c.Badges[i+1:]...)
			c.original.Badges = append(c.original.Badges[:i], c.original.Badges[i+1:]...)
		}
	default:
		if strings.HasPrefix(key, "VoteCategories-") {
			parts := strings.Split(key, "-")
			if len(parts) != 2 {
				return errors.New("bad vote index")
			}
			i, err := strconv.Atoi(parts[1])
			if err != nil {
				return errors.New("bad vote index")
			}
			if i < 0 || i >= len(c.VoteCategories) {
				return errors.New("bad vote index")
			}
			c.VoteCategories[i] = value
			c.original.VoteCategories[i] = value
		} else {
			return errors.New("no such key in config")
		}
	}
	return nil
}

func saveConfig() error {
	b, err := json.MarshalIndent(c.original, "", "	")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("config.json", b, 644)
}
