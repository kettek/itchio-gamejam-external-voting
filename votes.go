package main

import (
	"encoding/json"
	"errors"
	"strconv"

	"go.etcd.io/bbolt"
)

// Votes are the votes as stored in the DB.
type Votes struct {
	Audio      float64 `json:"audio"`
	Graphics   float64 `json:"graphics"`
	Innovation float64 `json:"innovation"`
	Gameplay   float64 `json:"gameplay"`
	Theme      float64 `json:"theme"`
}

func (v *Votes) Add(o Votes) {
	v.Audio += o.Audio
	v.Graphics += o.Graphics
	v.Innovation += o.Innovation
	v.Gameplay += o.Gameplay
	v.Theme += o.Theme
}

func (v *Votes) DivideBy(c float64) {
	if c == 0 {
		return
	}
	v.Audio /= c
	v.Graphics /= c
	v.Innovation /= c
	v.Gameplay /= c
	v.Theme /= c
}

// ErrMissingBucket is returned when a user bucket attempting to be accessed is missing.
var ErrMissingBucket = errors.New("missing bucket")

// ErrMissingGame is returned when there is no entry for the given game ID.
var ErrMissingGame = errors.New("missing game")

func setVotes(userID int, gameID int, votes Votes) error {
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(strconv.Itoa(userID)))
		if b == nil {
			return ErrMissingBucket
		}

		jsonBytes, err := json.Marshal(&votes)
		if err != nil {
			return err
		}

		err = b.Put([]byte(strconv.Itoa(gameID)), jsonBytes)
		return err
	})
	return err
}

func getVotes(userID int, gameID int) (Votes, error) {
	var votes Votes
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(strconv.Itoa(userID)))
		if b == nil {
			return ErrMissingBucket
		}

		jsonBytes := b.Get([]byte(strconv.Itoa(gameID)))
		if jsonBytes == nil {
			return ErrMissingGame
		}
		if err := json.Unmarshal(jsonBytes, &votes); err != nil {
			return err
		}
		return nil
	})
	return votes, err
}

func getFinalVotes(gameID int) (Votes, error) {
	var votes Votes
	var count float64
	err := db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			var userVotes Votes
			// Only accept valid entries.
			jsonBytes := b.Get([]byte(strconv.Itoa(gameID)))
			if jsonBytes == nil {
				return nil
			}
			if err := json.Unmarshal(jsonBytes, &userVotes); err != nil {
				return nil
			}

			votes.Add(userVotes)
			count++

			return nil
		})
	})

	votes.DivideBy(count)

	return votes, err
}
