package main

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"

	"go.etcd.io/bbolt"
)

// Votes are the votes as stored in the DB.
type Votes map[string]float64

// Add adds the passed vote's category values to this one.
func (votes Votes) Add(o Votes) {
	for k, v := range o {
		votes[k] += v
	}
}

// DivideBy divides all the vote categories by the given count.
func (votes Votes) DivideBy(c float64) {
	if c == 0 {
		return
	}
	for k := range votes {
		votes[k] /= c
	}
}

// ErrMissingBucket is returned when a user bucket attempting to be accessed is missing.
var ErrMissingBucket = errors.New("missing bucket")

// ErrMissingGame is returned when there is no entry for the given game ID.
var ErrMissingGame = errors.New("missing game")

// ErrOwnGame is returned when a user attempts to vote for their own game.
var ErrOwnGame = errors.New("cannot vote for own game")

func isOwnGame(user UserDetails, gameID int) bool {
	for _, entry := range entries.Games {
		if entry.ID != gameID {
			continue
		}
		if entry.Info.User.ID == user.ID {
			return true
		}
		for _, c := range entry.Contributors {
			// Name == DisplayName feels wrong to user, but contributors always use a user's DisplayName rather than the actual username. To prevent collisions if there is more than one user in a jam with the same display name, we also compare the contributor's URL to the user's URL. Note: We could probably just compare the URLs.
			if c.Name == user.DisplayName && c.URL == user.URL {
				return true
			}
		}
	}
	return false
}

func gameExists(gameID int) bool {
	for _, entry := range entries.Games {
		if entry.ID == gameID {
			return true
		}
	}
	return false
}

func setVotes(user UserDetails, gameID int, votes Votes) error {
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

		jsonBytes, err := json.Marshal(&votes)
		if err != nil {
			return err
		}

		err = b.Put([]byte("votes-"+strconv.Itoa(gameID)), jsonBytes)
		return err
	})
	return err
}

func getVotes(user UserDetails, gameID int) (Votes, error) {
	votes := make(Votes)
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(strconv.Itoa(user.ID)))
		if b == nil {
			return ErrMissingBucket
		}

		jsonBytes := b.Get([]byte("votes-" + strconv.Itoa(gameID)))
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
	votes := make(Votes)
	voteCounts := make(map[string]float64)
	err := db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			var userVotes Votes
			// Only accept valid entries.
			jsonBytes := b.Get([]byte("votes-" + strconv.Itoa(gameID)))
			if jsonBytes == nil {
				return nil
			}
			if err := json.Unmarshal(jsonBytes, &userVotes); err != nil {
				return nil
			}

			votes.Add(userVotes)
			for k, v := range votes {
				if v > 0 {
					voteCounts[k]++
				}
			}

			return nil
		})
	})

	for k, v := range voteCounts {
		votes[k] /= v
	}

	for k, v := range votes {
		votes[k] = math.Round(v*100) / 100
	}

	return votes, err
}
