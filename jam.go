package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

// loadJamInfo loads gamejam info if we have a game jam name and any of our game jam fields are empty.
func loadJamInfo() {
	if c.GameJam != "" && (c.GameJamID == 0 || c.GameJamName == "" || c.GameJamImage == "") {
		req, err := http.NewRequest("GET", fmt.Sprintf("https://itch.io/jam/%s/entries", c.GameJam), nil)
		if err != nil {
			log.Println("err", err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("err", err)
		} else if res.StatusCode == 200 {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Println("err", err)
			}

			// TODO: It would likely be better to use XML parsing for this, but regex is easy enough at the moment.

			// Get title
			if c.GameJamName == "" {
				r := regexp.MustCompile(`(?s)jam_title_header"><a href="[^\"]*">([^<]*)`)
				rs := r.FindStringSubmatch(string(body))
				if len(rs) == 2 {
					c.GameJamName = rs[1]
				}
			}
			// Get image
			if c.GameJamImage == "" {
				r := regexp.MustCompile(`(?s)cover_image"><img src="([^"]*)"`)
				rs := r.FindStringSubmatch(string(body))
				if len(rs) == 2 {
					c.GameJamImage = rs[1]
				}
			}
			// Get our game jam id
			if c.GameJamID == 0 {
				r := regexp.MustCompile(`(?s)"entries_url":"\\/jam\\/([0-9]+)\\/entries.json"`)
				rs := r.FindStringSubmatch(string(body))

				if len(rs) != 2 {
					panic(errors.New("could not acquire game jam id"))
				}
				c.GameJamID, err = strconv.Atoi(rs[1])
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
