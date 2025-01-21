package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/fogleman/gg"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/image/font"

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
			if err == ErrBadUser || err == ErrBadKey {
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
			http.Redirect(w, r, c.BaseURL+"/", http.StatusSeeOther)
		}
	})

	// Voting handling
	r.Get("/vote", func(w http.ResponseWriter, r *http.Request) {
		if c.VotingFinished {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting finished")
			return
		}
		if !c.VotingEnabled {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting disabled")
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

		for _, cat := range c.VoteCategories {
			v, err := getNum(cat)
			if err != nil {
				if err != ErrMissingQueryParam {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			} else {
				votes[cat] = v
			}
		}

		if err := setVotes(user.Details, id, votes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		// Return JSON with the user's current votes.
		b, err := json.Marshal(struct {
			ID    int                `json:"id"`
			Votes map[string]float64 `json:"votes"`
		}{
			ID:    id,
			Votes: votes,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})

	r.Get("/badge/*", func(w http.ResponseWriter, r *http.Request) {
		if !c.VotingFinished {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting not finished")
			return
		}
		part := r.URL.Path[len("/badge/"):]
		id, err := strconv.Atoi(part)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "bad id")
			return
		}
		w.Header().Set("Content-Type", "image/png")
		badges := getFinalBadges()
		for k, v := range badges {
			if s, ok := c.Badge.Rewrites[k]; ok {
				k = s
			}
			for _, v2 := range v {
				if v2 == id {
					w.Header().Set("Content-Type", "image/png")
					w.Write(generateBadge(k))
					return
				}
			}
		}
		dc := gg.NewContext(1, 1)
		dc.EncodePNG(w)
	})

	r.Get("/tags/*", func(w http.ResponseWriter, r *http.Request) {
		if !c.VotingFinished {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting not finished")
			return
		}
		part := r.URL.Path[len("/tags/"):]
		id, err := strconv.Atoi(part)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "bad id")
			return
		}
		// big ol' FIXME: Cache this crap!
		tags := getFinalTags()
		var tagStrings []string
		for k, v := range tags {
			for _, v2 := range v {
				if v2 == id {
					tagStrings = append(tagStrings, k)
				}
			}
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(generateTags(tagStrings))
	})

	r.Get("/cats/*", func(w http.ResponseWriter, r *http.Request) {
		if !c.VotingFinished {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting not finished")
			return
		}
		part := r.URL.Path[len("/cats/"):]
		id, err := strconv.Atoi(part)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "bad id")
			return
		}
		// big ol' FIXME: Cache this crap!
		cats := getTopCats()
		var myCats []CatResult
		for _, cat := range cats {
			if cat.id == id {
				myCats = append(myCats, cat)
			}
		}
		sort.Sort(CatResults(myCats))
		var catStrings []string
		for _, cat := range myCats {
			catStrings = append(catStrings, cat.cat)
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(generateCats(catStrings))
	})

	// Badge handling
	r.Get("/badge", func(w http.ResponseWriter, r *http.Request) {
		if c.VotingFinished {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting finished")
			return
		}
		if !c.VotingEnabled {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting disabled")
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

		badges, err := getBadges(user.Details, id)
		if err != nil {
			if err != ErrMissingGame {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, err.Error())
				return
			}
		}

		badge := q.Get("badge")

		var has bool
		for _, badge2 := range c.Badges {
			if badge2 == badge {
				has = true
				break
			}
		}
		if !has {
			// FIXME: Replace with proper err
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "bad badge")
			return
		}
		badges[badge] = !badges[badge]

		returnBadges, err := processBadge(user.Details, id, badges)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		// Return JSON with the user's current badges.
		b, err := json.Marshal(returnBadges)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})

	// Tag handling
	r.Get("/tag", func(w http.ResponseWriter, r *http.Request) {
		if c.VotingFinished {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting finished")
			return
		}
		if !c.VotingEnabled {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "voting disabled")
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

		tags, err := getTags(user.Details, id)
		if err != nil {
			if err != ErrMissingGame {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, err.Error())
				return
			}
		}

		tag := q.Get("tag")

		var has bool
		for _, tag2 := range c.Tags {
			if tag2 == tag {
				has = true
				break
			}
		}
		if !has {
			// FIXME: Replace with proper err
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "bad tag")
			return
		}

		tags[tag] = !tags[tag]
		if err := setTags(user.Details, id, tags); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		// Return JSON with the user's current tags.
		b, err := json.Marshal(struct {
			ID   int             `json:"id"`
			Tags map[string]bool `json:"tags"`
		}{
			ID:   id,
			Tags: tags,
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
			err := setConfig(k, v[0])
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Println(err)
				return
			}
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

		http.Redirect(w, r, c.BaseURL+"/", http.StatusSeeOther)
	})

	// Static file serving
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/*", fileServer)

}

// FIXME: Move this.
var badgeImage image.Image
var tagImage image.Image
var catImage image.Image
var fontFace font.Face

func loadCustom() {
	if c.Font != "" {
		stat, err := os.Stat(path.Join("custom", c.Font))
		if os.IsExist(err) || stat != nil {
			ff, err := gg.LoadFontFace(path.Join("custom", c.Font), c.FontSize)
			if err != nil {
				panic(err)
			}
			fontFace = ff
		}
	}
	b, err := os.ReadFile("custom/badge.png")
	if err != nil {
		fmt.Println("badge", err)
	}
	badgeImage, _, err = image.Decode(bytes.NewReader(b))
	if err != nil {
		fmt.Println("badge", err)
	}
	b, err = os.ReadFile("custom/tag.png")
	if err != nil {
		fmt.Println("badge", err)
	}
	tagImage, _, err = image.Decode(bytes.NewReader(b))
	if err != nil {
		fmt.Println("badge", err)
	}
	b, err = os.ReadFile("custom/cat.png")
	if err != nil {
		fmt.Println("cat", err)
	}
	catImage, _, err = image.Decode(bytes.NewReader(b))
	if err != nil {
		fmt.Println("cat", err)
	}
}

var badgeLock sync.Mutex
var badgeImages map[string][]byte = make(map[string][]byte)

func generateBadge(text string) []byte {
	badgeLock.Lock()
	defer badgeLock.Unlock()

	if r, ok := badgeImages[text]; ok {
		return r
	}

	if text == "" {
		dc := gg.NewContext(1, 1)
		w := new(bytes.Buffer)
		dc.EncodePNG(w)
		badgeImages[text] = w.Bytes()
		return badgeImages[text]
	}
	dc := gg.NewContext(c.Badge.Width, c.Badge.Height)
	if fontFace != nil {
		dc.SetFontFace(fontFace)
	}

	dc.DrawImage(badgeImage, 0, 0)

	dc.SetRGB(0, 0, 0)
	dc.DrawStringAnchored(text, float64(c.Badge.TextX), float64(c.Badge.TextY), 0.5, 0.5)

	w := new(bytes.Buffer)
	dc.EncodePNG(w)
	badgeImages[text] = w.Bytes()
	return badgeImages[text]
}

var tagLock sync.Mutex
var tagImages map[string][]byte = make(map[string][]byte)

func generateTags(tags []string) []byte {
	tagLock.Lock()
	defer tagLock.Unlock()

	tagKey := strings.Join(tags, "+")

	if r, ok := tagImages[tagKey]; ok {
		return r
	}

	if len(tags) == 0 {
		dc := gg.NewContext(1, 1)
		w := new(bytes.Buffer)
		dc.EncodePNG(w)
		tagImages[tagKey] = w.Bytes()
		return tagImages[tagKey]
	}

	dc := gg.NewContext(c.Tag.Width*len(tags)+c.Tag.Width/8*(len(tags)-1), c.Tag.Height)
	if fontFace != nil {
		dc.SetFontFace(fontFace)
	}
	x := 0
	for _, tag := range tags {
		if s, ok := c.Badge.Rewrites[tag]; ok {
			tag = s
		}

		dc.DrawImage(tagImage, x, 0)
		dc.SetRGB(0, 0, 0)
		dc.DrawStringAnchored(tag, float64(c.Tag.TextX)+float64(x), float64(c.Tag.TextY), 0.5, 0.5)
		x += c.Tag.Width + c.Tag.Width/8
	}

	w := new(bytes.Buffer)

	dc.EncodePNG(w)
	tagImages[tagKey] = w.Bytes()
	return tagImages[tagKey]
}

var catLock sync.Mutex
var catImages map[string][]byte = make(map[string][]byte)

func generateCats(cats []string) []byte {
	catLock.Lock()
	defer catLock.Unlock()

	catKey := strings.Join(cats, "+")

	if r, ok := catImages[catKey]; ok {
		return r
	}

	if len(cats) == 0 {
		dc := gg.NewContext(1, 1)
		w := new(bytes.Buffer)
		dc.EncodePNG(w)
		catImages[catKey] = w.Bytes()
		return catImages[catKey]
	}

	dc := gg.NewContext(c.Cat.Width*len(cats)+c.Cat.Width/8*(len(cats)-1), c.Cat.Height)
	if fontFace != nil {
		dc.SetFontFace(fontFace)
	}

	x := 0
	for _, cat := range cats {
		if s, ok := c.Cat.Rewrites[cat]; ok {
			cat = s
		}

		dc.DrawImage(catImage, x, 0)
		dc.SetRGB(0, 0, 0)
		dc.DrawStringAnchored(cat, float64(c.Cat.TextX)+float64(x), float64(c.Cat.TextY), 0.5, 0.5)
		x += c.Cat.Width + c.Cat.Width/8
	}

	w := new(bytes.Buffer)

	dc.EncodePNG(w)
	catImages[catKey] = w.Bytes()
	return catImages[catKey]
}
