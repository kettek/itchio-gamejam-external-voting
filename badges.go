package main

import (
	"bytes"
	"encoding/json"
	"strconv"

	"go.etcd.io/bbolt"
)

// Badges are the badges as stored in the DB.
type Badges map[string]bool

// processBadge iterates through all of the player's game badge votes and removes any badges that conflict with the passed in badge.
func processBadge(user UserDetails, gameID int, badges Badges) (map[int]Badges, error) {
	if isOwnGame(user, gameID) {
		return nil, ErrOwnGame
	}
	if !gameExists(gameID) {
		return nil, ErrMissingGame
	}
	returnBadges := make(map[int]Badges)
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(strconv.Itoa(user.ID)))
		if b == nil {
			return ErrMissingBucket
		}
		c := b.Cursor()

		// Iterate all buckets of b with prefix badges-
		prefix := []byte("badges-")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			id, err := strconv.Atoi(string(k[len(prefix):]))
			if err != nil {
				return err
			}
			var badges2 Badges
			if err := json.Unmarshal(v, &badges2); err != nil {
				return err
			}
			changed := false
			if id != gameID {
				for k, v := range badges {
					if !v {
						continue
					}
					if v := badges2[k]; v {
						delete(badges2, k)
						changed = true
					}
				}
			}
			if changed {
				jsonBytes, err := json.Marshal(&badges2)
				if err != nil {
					return err
				}
				if err := b.Put(k, jsonBytes); err != nil {
					return err
				}
				returnBadges[id] = badges2
			}
		}

		jsonBytes, err := json.Marshal(&badges)
		if err != nil {
			return err
		}
		err = b.Put([]byte("badges-"+strconv.Itoa(gameID)), jsonBytes)

		returnBadges[gameID] = badges

		return err
	})
	return returnBadges, err
}

func getBadges(user UserDetails, gameID int) (Badges, error) {
	badges := Badges{}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(strconv.Itoa(user.ID)))
		if b == nil {
			return ErrMissingBucket
		}

		jsonBytes := b.Get([]byte("badges-" + strconv.Itoa(gameID)))
		if jsonBytes == nil {
			return ErrMissingGame
		}
		if err := json.Unmarshal(jsonBytes, &badges); err != nil {
			return err
		}
		return nil
	})
	return badges, err
}
