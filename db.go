package main

import (
	"fmt"
	"os"

	bolt "go.etcd.io/bbolt"
)

func loadDB() *bolt.DB {
	dbPath := c.DBRoot
	if dbPath == "" {
		dbPath = "db"
	}

	dbName := c.GameJam
	if dbName == "" {
		dbName = "default"
	}

	if err := os.MkdirAll(dbPath, 0770); err != nil {
		panic(err)
	}

	db, err := bolt.Open(fmt.Sprintf("db/%s", dbName), 0666, nil)
	if err != nil {
		panic(err)
	}
	return db
}
