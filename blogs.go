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

// BlogMeta is the metadata for a blog post in the database
type BlogMeta struct {
	ID    string `storm:"unique"`
	Title string `storm:"index"`
	Date  string `storm:"index"`
	Desc  string `storm:"index"`
}

// BlogContent is the content of a blog post in the database
type BlogContent struct {
	ID      string `storm:"unique"`
	Content string `storm:"index"`
}

// BlogIdGen generates a unique id for a blog post
func BlogIdGen(title, date string) string {
	title = strings.ReplaceAll(title, " ", "-")
	return filepath.Join(date, url.PathEscape(title))
}

// SaveBlog saves a blog post to the database with arguments
func SaveBlog(title, desc, body string, date ...string) error {
	if len(date) == 0 {
		date = append(date, time.Now().Format("2006-01-02"))
	}
	id := BlogIdGen(title, date[0])

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

// GetBlog retrieves a blog post from the database by id and returns the metadata and content
// as BlogMeta and BlogContent respectively. Returns an error if the blog post is not found.
func GetBlog(id string) (meta BlogMeta, content BlogContent, err error) {
	err = DB.One("ID", id, &content)
	if err != nil {
		return
	}
	err = DB.One("ID", id, &meta)
	return
}

// GetBlogs retrieves all blog posts from the database and returns them as a slice of BlogMeta.
// Returns an error if the database query fails.
func GetBlogs() ([]BlogMeta, error) {
	var blogs []BlogMeta
	// err := DB.All(&blogs)
	err := DB.AllByIndex("Date", &blogs)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(blogs); i++ {
		for j := i + 1; j < len(blogs); j++ {
			if blogs[i].Date < blogs[j].Date {
				blogs[i], blogs[j] = blogs[j], blogs[i]
			}
		}
	}
	return blogs, err
}

// BlogHandler handles requests to the blog page and individual blog posts
func BlogHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/blog/")
	if path == "" {
		blogs, err := GetBlogs()
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		d, err := GetPageData("blogs")
		if CheckHttpErr(err, w, r, 500) {
			return
		}
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

// BlogManageList handles the /manage/blog/ route for listing all blog posts
func BlogManageList(w http.ResponseWriter, r *http.Request) {
	blogs, err := GetBlogs()
	if CheckHttpErr(err, w, r, 500) {
		return
	}
	opts := make([]UrlOpt, len(blogs)+1)
	opts[0] = UrlOpt{Name: "Add new blog", URL: "/manage/blog/?id=new"}
	for i, b := range blogs {
		opts[i+1] = UrlOpt{Name: b.Title, URL: "/manage/blog/?id=" + b.ID}
	}
	d := template.Data("Manage blogs", "List of blogs")
	d.Set("Options", opts)
	err = template.Use(w, r, "manage", d)
	CheckHttpErr(err, w, r, 500)
	return
}

// BlogManager handles the /manage/blog/ route for managing blog posts
func BlogManager(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/manage/blog/" {
		HttpErr(w, r, 404)
		return
	}
	if r.Method == "GET" {
		id := r.URL.Query().Get("id")
		if id == "" {
			BlogManageList(w, r)
			return
		}
		if id == "new" {
			var data = template.Data("New blog", "Create a new blog")
			data.Set("BlogUrl", "/manage/blog/")
			data.Set("BlogID", "new")
			data.Set("BlogDesc", "")
			data.Set("BlogContent", "")
			err := template.Use(w, r, "edit", data)
			CheckHttpErr(err, w, r, 500)
			return
		}
		meta, content, err := GetBlog(id)
		if CheckHttpErr(err, w, r, 404) {
			return
		}
		data := template.Data("Edit blog "+meta.Title, "Make changes to the content or description")
		data.Set("BlogUrl", "/manage/blog/")
		data.Set("BlogID", meta.ID)
		data.Set("BlogDesc", meta.Desc)
		data.Set("BlogContent", content.Content)
		err = template.Use(w, r, "edit", data)
		CheckHttpErr(err, w, r, 500)
		return
	}
	if r.Method == "POST" {
		err := r.ParseForm()
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		id := r.Form.Get("id")
		title := r.Form.Get("title")
		desc := r.Form.Get("desc")
		body := r.Form.Get("content")
		date := r.Form.Get("date")
		if id == "" || desc == "" || body == "" || (id == "new" && title == "") {
			HttpErr(w, r, 400)
			return
		}
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		if id == "new" {
			err = SaveBlog(title, desc, body, date)
			CheckHttpErr(err, w, r, 500)
			http.Redirect(w, r, "/manage/blog/", 302)
			return
		}
		meta, content, err := GetBlog(id)
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		if meta.Desc != desc {
			meta.Desc = desc
			err = DB.Update(&meta)
			if CheckHttpErr(err, w, r, 500) {
				return
			}
		}
		if content.Content != body {
			content.Content = body
			err = DB.Update(&content)
			if CheckHttpErr(err, w, r, 500) {
				return
			}
		}
		http.Redirect(w, r, "/manage/blog/", 302)
		return
	}
}
