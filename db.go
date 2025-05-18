package main

import (
	"encoding/json"
	"fmt"
	"git.sophuwu.com/myweb/config"
	"git.sophuwu.com/myweb/template"
	"github.com/asdine/storm/v3"
	"go.etcd.io/bbolt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
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

// AddRequiredData looks for important data types in the db
// and adds a placeholder if they don't exist. This is useful
// for the first run of the server to prevent errors.
func AddRequiredData() {
	d, err := GetPageData("index")
	if err != nil || d == nil {
		_ = SetPageData("index", template.HTMLDataMap{
			"Title":       "The title of the page",
			"Description": "The description for meta tags",
			"AboutText": []string{
				"Paragraph 1",
			},
			"ImagePath": "/path/to/pic.jpg",
			"Profiles": []Profile{
				{"char in iconfont", "name of url", "URL", "username"},
			},
			"Content": "<p>HTML content</p>",
		})
	}
	d, err = GetPageData("blogs")
	if err != nil || d == nil {
		_ = SetPageData("blogs", template.HTMLDataMap{
			"Title":       "The title of the page",
			"Description": "The description for meta tags",
		})
	}
	d, err = GetPageData("anims")
	if err != nil || d == nil {
		_ = SetPageData("anims", template.HTMLDataMap{
			"Title":       "The title of the page",
			"Description": "The description for meta tags",
		})
	}
}

// GetPageData returns a map of page metadata and data
// used for index/list pages which don't have a separate
// data source.
func GetPageData(page string) (template.HTMLDataMap, error) {
	var d template.HTMLDataMap
	err := DB.Get("pages", page, &d)
	return d, err
}

// SetPageData writes a HTMLDataMap to the db for persistent
// storage of page data without its own data source.
func SetPageData(page string, data template.HTMLDataMap) error {
	return DB.Set("pages", page, data)
}

// EditIndex handles the /manage/edit/ route for editing the
// index page's data and metadata.
func EditIndex(w http.ResponseWriter, r *http.Request) {
	var err error
	page := r.URL.Query().Get("page")
	if page == "" {
		HttpErr(w, r, 404)
	}
	if r.Method == "GET" && r.URL.Path == "/manage/edit/" {
		var d template.HTMLDataMap
		d, err = GetPageData(page)
		if CheckHttpErr(err, w, r, 404) {
			return
		}
		var b []byte
		b, err = json.MarshalIndent(d, "", "  ")
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		bb := string(b)
		data := template.Data("Edit "+page, "Edit the page's data and metadata")
		data.Set("Data", bb)
		data.Set("EditUrl", "/manage/edit/save?page="+page)
		err = template.Use(w, r, "edit", data)
		CheckHttpErr(err, w, r, 500)
		return
	} else if r.Method == "POST" && r.URL.Path == "/manage/edit/save" {
		err = r.ParseForm()
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		var bb string
		bb = r.Form.Get("data")
		if bb == "" {
			HttpErr(w, r, 400)
			return
		}
		var d template.HTMLDataMap
		err = json.Unmarshal([]byte(bb), &d)
		if CheckHttpErr(err, w, r, 400) {
			return
		}
		err = SetPageData(page, d)
		if CheckHttpErr(err, w, r, 400) {
			return
		}
		http.Redirect(w, r, "/", 302)
		return
	}
	HttpErr(w, r, 405)
}

// UrlOpt is a struct for the options on the manage page.
type UrlOpt struct{ Name, URL string }

// ManagerHandler handles the /manage/ route for managing
// the website. It displays a list of options for managing
// the website's content. And forwards to the appropriate
// handler for each option.
func ManagerHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/manage/" {
		data := template.Data("Manage", config.URL)
		data["Options"] = []UrlOpt{
			{"Edit index", "/manage/edit/?page=index"},
			{"Edit blog list meta", "/manage/edit/?page=blogs"},
			{"Edit animation meta", "/manage/edit/?page=anims"},
			{"Upload media", "/manage/media/"},
			{"Delete media", "/manage/delete/media/"},
			{"Manage blogs", "/manage/blog/"},
			{"Manage Animations", "/manage/animation/"},
			{"Backup", "/manage/backup/"},
		}
		err := template.Use(w, r, "manage", data)
		CheckHttpErr(err, w, r, 500)
		return
	}
	if r.URL.Path == "/manage/edit/" || r.URL.Path == "/manage/edit/save" {
		EditIndex(w, r)
		return
	}
	if r.URL.Path == "/manage/media/" {
		ManageMedia(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/manage/delete/media/") {
		DeleteMedia(w, r)
		return
	}
	if r.URL.Path == "/manage/animation/" {
		AnimManager(w, r)
		return
	}
	if r.URL.Path == "/manage/blog/" {
		BlogManager(w, r)
		return
	}
	if r.URL.Path == "/manage/backup/" {
		BackerUpper(w, r)
		return
	}
	HttpErr(w, r, 404)
}

// BackerUpper is a handler to download the database file.
func BackerUpper(w http.ResponseWriter, r *http.Request) {
	err := DB.Bolt.View(func(tx *bbolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-data.db.%s.bak"`, filepath.Base(config.URL), time.Now().Format("2006-01-02")))
		w.Header().Set("Content-Length", fmt.Sprint(int(tx.Size())))
		_, err := tx.WriteTo(w)
		return err
	})
	if CheckHttpErr(err, w, r, 500) {
		return
	}
}
