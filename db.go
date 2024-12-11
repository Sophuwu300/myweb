package main

import (
	"encoding/json"
	"github.com/asdine/storm/v3"
	"log"
	"net/http"
	"sophuwu.site/myweb/config"
	"sophuwu.site/myweb/template"
	"strings"
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

func EditIndex(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method == "GET" && r.URL.Path == "/manage/edit/" {
		var d template.HTMLDataMap
		d, err = GetPageData("index")
		var b []byte
		b, err = json.MarshalIndent(d, "", "  ")
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		bb := string(b)
		data := template.Data("Edit index", "Edit the index page")
		data.Set("Data", bb)
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
		err = SetPageData("index", d)
		if CheckHttpErr(err, w, r, 400) {
			return
		}
		http.Redirect(w, r, "/", 302)
		return
	}
	HttpErr(w, r, 405)
}

type UrlOpt struct{ Name, URL string }

func ManagerHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/manage/" {
		data := template.Data("Manage", config.URL)
		data["Options"] = []UrlOpt{
			{"Edit index", "/manage/edit/"},
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
	HttpErr(w, r, 404)
}
