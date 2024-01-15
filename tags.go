package main

import (
	"encoding/json"
	"strconv"

	"go.etcd.io/bbolt"
)

// Tags are the tags as stored in the DB.
type Tags map[string]bool

func setTags(user UserDetails, gameID int, tags Tags) error {
	if isOwnGame(user, gameID) {
		return ErrOwnGame
	}
	if !gameExists(gameID) {
		return ErrMissingGame
	}

	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(strconv.Itoa(user.ID)))
		if b == nil {
			return ErrMissingBucket
		}

		jsonBytes, err := json.Marshal(&tags)
		if err != nil {
			return err
		}

		err = b.Put([]byte(strconv.Itoa(gameID)+"-tags"), jsonBytes)
		return err
	})
	return err
}

func getTags(user UserDetails, gameID int) (Tags, error) {
	tags := Tags{}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(strconv.Itoa(user.ID)))
		if b == nil {
			return ErrMissingBucket
		}

		jsonBytes := b.Get([]byte(strconv.Itoa(gameID) + "-tags"))
		if jsonBytes == nil {
			return ErrMissingGame
		}
		if err := json.Unmarshal(jsonBytes, &tags); err != nil {
			return err
		}
		return nil
	})
	return tags, err
}
