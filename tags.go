package main

import (
	"bytes"
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

		err = b.Put([]byte("tags-"+strconv.Itoa(gameID)), jsonBytes)
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

		jsonBytes := b.Get([]byte("tags-" + strconv.Itoa(gameID)))
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

type TagResults map[string][]int

func getFinalTags() TagResults {
	results := make(TagResults)
	counts := make(map[string]map[int]int)
	db.View(func(tx *bbolt.Tx) error {
		tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			c := b.Cursor()
			prefix := []byte("tags-")
			for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
				id, err := strconv.Atoi(string(k[len(prefix):]))
				if err != nil {
					return err
				}
				tags := Tags{}
				if err := json.Unmarshal(v, &tags); err != nil {
					return err
				}
				for k, v := range tags {
					if v {
						if _, ok := counts[k]; !ok {
							counts[k] = make(map[int]int)
						}
						counts[k][id]++
					}
				}
			}
			return nil
		})
		return nil
	})

	// Collect the highest rated counts.
	for k, v := range counts {
		for _, count := range v {
			for id2, count2 := range v {
				if count2 < count {
					delete(counts[k], id2)
				}
			}
		}
	}

	for k, v := range counts {
		for id := range v {
			results[k] = append(results[k], id)
		}
	}

	return results
}
