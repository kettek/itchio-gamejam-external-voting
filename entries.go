package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// User is the result of the me API endpoint.
type User struct {
	Details UserDetails `json:"user"`
}

// UserDetails is the actual user details.
type UserDetails struct {
	ID          int
	ImageURL    string `json:"cover_url"`
	DisplayName string `json:"display_name"`
	UserName    string `json:"username"`
	URL         string `json:"url"`
}

// Entries is the main entries struct returned from itch.io
type Entries struct {
	Games []GameEntry `json:"jam_games"`
}

// GameEntry is an individual gamejam entry.
type GameEntry struct {
	ID           int
	Info         GameInfo `json:"game"`
	URL          string
	Contributors []UserInfo
}

// GameInfo is the specific details for a gamejam entry.
type GameInfo struct {
	ID         int
	Platforms  []string
	Cover      string
	Title      string
	ShortText  string `json:"short_text"`
	CoverColor string `json:"cover_color"`
	URL        string
	User       UserInfo
}

// UserInfo is the user information for a given game entry.
type UserInfo struct {
	ID   int
	URL  string
	Name string
}

func getEntries(id int) Entries {
	url := fmt.Sprintf("https://itch.io/jam/%d/entries.json", id)

	cl := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	res, err := cl.Do(req)
	if err != nil {
		panic(err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &entries)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Read in %d entries :)\n", len(entries.Games))

	return entries
}
