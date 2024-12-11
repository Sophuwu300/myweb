package main

import (
	_ "github.com/asdine/storm/v3"
	_ "go.etcd.io/bbolt"
	"net/http"
	"net/url"
	"path/filepath"
	"sophuwu.site/myweb/template"
	"strings"
	"time"
)

type BlogMeta struct {
	ID    string `storm:"unique"`
	Title string `storm:"index"`
	Date  string `storm:"index"`
	Desc  string `storm:"index"`
}

type BlogContent struct {
	ID      string `storm:"unique"`
	Content string `storm:"index"`
}

func IdGen(title, date string) string {
	title = strings.ReplaceAll(title, " ", "-")
	return filepath.Join(date, url.PathEscape(title))
}

func SaveBlog(title, desc, body string, date ...string) error {
	if len(date) == 0 {
		date = append(date, time.Now().Format("2006-01-02"))
	}
	id := IdGen(title, date[0])

	err := DB.Save(&BlogContent{
		ID:      id,
		Content: body,
	})
	if err != nil {
		return err
	}

	blg := BlogMeta{
		ID:    id,
		Title: title,
		Date:  date[0],
		Desc:  desc,
	}
	err = DB.Save(&blg)
	return err
}

func GetBlog(id string) (meta BlogMeta, content BlogContent, err error) {
	err = DB.One("ID", id, &content)
	if err != nil {
		return
	}
	err = DB.One("ID", id, &meta)
	return
}

func SortBlogsDate(blogs []BlogMeta) []BlogMeta {
	for i := 0; i < len(blogs); i++ {
		for j := i + 1; j < len(blogs); j++ {
			if blogs[i].Date < blogs[j].Date {
				blogs[i], blogs[j] = blogs[j], blogs[i]
			}
		}
	}
	return blogs
}

func GetBlogs() ([]BlogMeta, error) {
	var blogs []BlogMeta
	// err := DB.All(&blogs)
	err := DB.AllByIndex("Date", &blogs)
	if err != nil {
		return nil, err
	}
	blogs = SortBlogsDate(blogs)
	return blogs, err
}

func BlogHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/blog/")
	if path == "" {
		blogs, err := GetBlogs()
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		d := template.Data("Sophie's Blogs", "I sometimes write blogs about random things that I find interesting. Here you can read all my posts about various things I found interesting at some point.")
		d["blogs"] = []BlogMeta(blogs)
		d.Set("NoBlogs", len(blogs))

		err = template.Use(w, r, "blogs", d)
		CheckHttpErr(err, w, r, 500)
		return
	}
	meta, content, err := GetBlog(path)
	if CheckHttpErr(err, w, r, 404) {
		return
	}
	data := template.Data(meta.Title, meta.Desc)
	data.Set("Date", meta.Date)
	data.SetHTML(content.Content)
	err = template.Use(w, r, "blog", data)
	CheckHttpErr(err, w, r, 500)
}
