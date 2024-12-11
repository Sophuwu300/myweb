package main

import (
	"github.com/asdine/storm/v3"
	"go.etcd.io/bbolt"
	"log"
	"sophuwu.site/myweb/config"
	"time"
)

var DB *storm.DB

func OpenDB() {
	db, err := storm.Open(config.DBPath, storm.BoltOptions(0660, &bbolt.Options{Timeout: time.Second}))
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	DB = db
}

func CloseDB() {
	err := DB.Close()
	if err != nil {
		log.Println(err)
	}
}
