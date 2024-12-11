package main

import (
	"fmt"
	"github.com/asdine/storm/v3"
	"net/http"
	"strings"
	"time"
)

type BlogMeta struct {
	ID    string `storm:"id"`
	Title string
	Date  string `storm:"index"`
	Desc  string
}

func NewBlog(title, desc, body string, date ...string) error {
	if len(date) == 0 {
		date = append(date, time.Now().Format("2006-01-02"))
	}
	id := Sha1Base64(title, date[0])

	exists, err := DB.KeyExists("BlogContent", id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("blog already exists")
	}

	err = DB.Set("BlogContent", id, body)
	if err != nil {
		return err
	}

	blg := BlogMeta{
		ID:    id,
		Title: title,
		Date:  date[0],
		Desc:  desc,
	}
	return DB.Save(&blg)
}

func GetBlog(id string) (meta BlogMeta, content string, err error) {
	err = DB.Get("BlogContent", id, &content)
	if err != nil {
		return
	}
	err = DB.One("ID", id, &meta)
	return
}

func GetBlogs() ([]BlogMeta, error) {
	var blogs []BlogMeta
	err := DB.AllByIndex("Date", &blogs, storm.Limit(10), storm.Reverse())
	return blogs, err
}

func BlogHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/blog/")
	if path == "" {
		blogs, err := GetBlogs()
		if CheckHttpErr(err, w, r, 500) {
			return
		}

		// 		"Title": "Sophie's Blogs",
		// "Desc":  "I sometimes write blogs about random things that I find interesting. Here you can read all my posts about various things I found interesting at some point.",
		// "Blogs": blogs,
		err = Templates.Use(w, "blogs")
		CheckHttpErr(err, w, r, 500)
		return
	}
}
