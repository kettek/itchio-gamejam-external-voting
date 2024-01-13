package main

import (
	"encoding/json"
	"strconv"

	"go.etcd.io/bbolt"
)

// UserBadges are the badges as stored in the DB.
type UserBadges map[string]Badges

type Badges struct {
	Uniques    []string
	NonUniques []string
}

func (b Badges) Add(o string, unique bool) {
	if unique {
		b.Uniques = append(b.Uniques, o)
	} else {
		b.NonUniques = append(b.NonUniques, o)
	}
}

func setBadges(user UserDetails, gameID int, badges Badges) error {
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

		jsonBytes, err := json.Marshal(&badges)
		if err != nil {
			return err
		}

		err = b.Put([]byte(strconv.Itoa(gameID)+"-badges"), jsonBytes)
		return err
	})
	return err
}

func getBadges(user UserDetails, gameID int) (Badges, error) {
	badges := Badges{}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(strconv.Itoa(user.ID)))
		if b == nil {
			return ErrMissingBucket
		}

		jsonBytes := b.Get([]byte(strconv.Itoa(gameID) + "-badges"))
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
