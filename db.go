package main

import (
	"github.com/asdine/storm/v3"
	"log"
	"sophuwu.site/myweb/config"
	"sophuwu.site/myweb/template"
)

var DB *storm.DB

func OpenDB() {
	db, err := storm.Open(config.DBPath)
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

func GetPageData(page string) (template.HTMLDataMap, error) {
	var d template.HTMLDataMap
	err := DB.Get("pages", page, &d)
	return d, err
}

func SetPageData(page string, data template.HTMLDataMap) error {
	return DB.Set("pages", page, data)
}
