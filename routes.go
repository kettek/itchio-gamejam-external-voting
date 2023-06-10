package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	bolt "go.etcd.io/bbolt"

	"gitea.com/go-chi/session"
)

// ErrMissingQueryParam is returned when a vote category is not defined in a query when checked for.
var ErrMissingQueryParam = errors.New("missing query param")

// ErrBadUser is returned when a given key cannot be used to get a user.
var ErrBadUser = errors.New("no such user for key")

// ErrBadKey is returned when a given key is invalid for itch.io
var ErrBadKey = errors.New("invalid key")

// ErrNoUser is returned when a given key has no user available.
var ErrNoUser = errors.New("no such user for key")

func getUserFromKey(key string) (User, error) {
	var user User
	if key != "" {
		userReq, err := http.NewRequest("GET", fmt.Sprintf("https://itch.io/api/1/%s/me", key), nil)
		if err != nil {
			return user, errors.Join(ErrBadUser, err)
		}
		res, err := http.DefaultClient.Do(userReq)
		if err != nil {
			return user, errors.Join(ErrBadUser, err)
		}
		if res.StatusCode == 200 {
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return user, errors.Join(ErrBadUser, err)
			}
			// Check for itch.io error first.
			itchErr := struct {
				Errors []string
			}{}
			if err := json.Unmarshal(resBody, &itchErr); err == nil {
				if len(itchErr.Errors) > 0 {
					return user, errors.Join(ErrBadKey, err)
				}
			}
			// Now try to read the user!
			if err := json.Unmarshal(resBody, &user); err != nil {
				return user, errors.Join(ErrBadUser, err)
			}
			return user, nil
		}
	}
	return user, ErrNoUser
}

func getSessionKey(r *http.Request) string {
	s := session.GetSession(r)
	var key string
	switch v := s.Get("key").(type) {
	case string:
		key = v
	}
	return key
}

func isAdmin(user User) bool {
	for _, a := range c.Admins {
		if a.ID != 0 && a.ID == user.Details.ID {
			return true
		}
		if a.URL != "" && a.URL == user.Details.URL {
			return true
		}
		if a.Name != "" && a.Name == user.Details.DisplayName {
			return true
		}
	}
	return false
}

func setupRoutes() {
	// Index and voting handling.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		var authed, admin bool
		key := getSessionKey(r)
		user, err := getUserFromKey(key)
		if err != nil {
			// Clear out the key, as the auth must've been revoked.
			if err == ErrNoUser || err == ErrBadKey {
				session.GetSession(r).Set("key", nil)
			}
		} else {
			authed = true
			// FIXME: This should be done during the initial auth stage.
			// Ensure the user has a bucket.
			if err := db.Update(func(tx *bolt.Tx) error {
				_, err := tx.CreateBucketIfNotExists([]byte(strconv.Itoa(user.Details.ID)))
				return err
			}); err != nil {
				log.Println("err", err)
			}
			// Set as admin if the user is in the admins slice.
			admin = isAdmin(user)
		}

		if err := templates.ExecuteTemplate(w, "index.gohtml", struct {
			Name           string
			Image          string
			Entries        Entries
			Authed         bool
			Admin          bool
			User           User
			VotingEnabled  bool
			VotingFinished bool
			Config         Config
		}{
			Name:           c.GameJamName,
			Image:          c.GameJamImage,
			VotingEnabled:  c.VotingEnabled,
			VotingFinished: c.VotingFinished,
			Entries:        entries,
			Authed:         authed,
			Admin:          admin,
			User:           user,
			Config:         c,
		}); err != nil {
			panic(err)
		}
	})

	// Login/logout and OAuth handling
	r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			templates.ExecuteTemplate(w, "auth.gohtml", nil)
		} else {
			s := session.GetSession(r)
			s.Set("key", key)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})

	// Voting handling
	r.Get("/vote", func(w http.ResponseWriter, r *http.Request) {
		if !c.VotingEnabled {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting disabled")
			return
		}
		if c.VotingFinished {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting finished")
			return
		}

		q := r.URL.Query()
		id, err := strconv.Atoi(q.Get("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "empty id")
			return
		}

		key := getSessionKey(r)
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "missing key")
			return
		}
		user, err := getUserFromKey(key)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, err.Error())
			return
		}

		// Get the previous votes for the given game.
		votes, err := getVotes(user.Details, id)
		if err != nil {
			if err != ErrMissingGame {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, err.Error())
				return
			}
		}

		getNum := func(which string) (float64, error) {
			s := q.Get(which)
			if s == "" {
				return 0, ErrMissingQueryParam
			}
			v, err := strconv.ParseFloat(s, 64)

			if err != nil {
				return 0, err
			}
			if v < 0 {
				return 0, nil
			}
			if v > 5 {
				return 5, nil
			}
			return v, nil
		}

		// TODO: Use a map or something.
		audio, err := getNum("audio")
		if err != nil {
			if err != ErrMissingQueryParam {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			votes.Audio = audio
		}
		graphics, err := getNum("graphics")
		if err != nil {
			if err != ErrMissingQueryParam {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			votes.Graphics = graphics
		}
		innovation, err := getNum("innovation")
		if err != nil {
			if err != ErrMissingQueryParam {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			votes.Innovation = innovation
		}
		gameplay, err := getNum("gameplay")
		if err != nil {
			if err != ErrMissingQueryParam {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			votes.Gameplay = gameplay
		}
		theme, err := getNum("theme")
		if err != nil {
			if err != ErrMissingQueryParam {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			votes.Theme = theme
		}

		if err := setVotes(user.Details, id, votes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		// Return JSON with the user's current votes.
		b, err := json.Marshal(struct {
			ID         int     `json:"id"`
			Audio      float64 `json:"audio"`
			Graphics   float64 `json:"graphics"`
			Innovation float64 `json:"innovation"`
			Gameplay   float64 `json:"gameplay"`
			Theme      float64 `json:"theme"`
		}{
			ID:         id,
			Audio:      votes.Audio,
			Graphics:   votes.Graphics,
			Innovation: votes.Innovation,
			Gameplay:   votes.Gameplay,
			Theme:      votes.Theme,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})

	// Admin
	r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		key := getSessionKey(r)
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "missing key")
			return
		}
		user, err := getUserFromKey(key)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, err.Error())
			return
		}

		if !isAdmin(user) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		for k, v := range q {
			setConfig(k, v[0])
		}
		if err := saveConfig(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		b, err := json.Marshal(c)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("https://itch.io/user/oauth?client_id=%s&scope=%s&response_type=token&redirect_uri=%s", c.ClientID, url.QueryEscape("profile:me"), url.QueryEscape(c.OAuthRedirect)), http.StatusSeeOther)
	})
	r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
		s := session.GetSession(r)
		s.Destroy(w, r)

		// TODO: If possible, somehow revoke auth from itch.io. They don't seem to have a proper endpoint for this...

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Static file serving
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/*", fileServer)

}
