package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sophuwu.site/myweb/template"
	"strings"
	"time"
)

type AnimInfo struct {
	ID    string `storm:"id"`
	Title string
	Date  string `storm:"index"`
	Desc  string
	Imgs  []string
	Vids  []string
}

func (a *AnimInfo) HasReqFields() bool {
	return a.Title != "" && a.Desc != "" && (len(a.Imgs)+len(a.Vids) > 0) && a.Date != "" && a.ID != ""
}

func GenAnimID(a AnimInfo) AnimInfo {
	md := md5.New()
	md.Write([]byte(time.Now().String() + a.Title))
	a.ID = base64.URLEncoding.EncodeToString(md.Sum(nil))
	if a.Date == "" {
		a.Date = time.Now().Format("2006-01-02")
	}
	return a
}

func GetAnim(id string) (AnimInfo, error) {
	var a AnimInfo
	err := DB.One("ID", id, &a)
	return a, err
}

func AnimSaveJson(js string) error {
	var a AnimInfo
	err := json.Unmarshal([]byte(js), &a)
	if err != nil {
		return err
	}
	return DB.Save(&a)
}

func AnimDelete(id string) error {
	return DB.DeleteStruct(&AnimInfo{ID: id})
}

func GetAnims() ([]AnimInfo, error) {
	var anims []AnimInfo
	err := DB.All(&anims)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(anims); i++ {
		for j := i + 1; j < len(anims); j++ {
			if anims[i].Date < anims[j].Date {
				anims[i], anims[j] = anims[j], anims[i]
			}
		}
	}
	return anims, nil
}

func AnimHandler(w http.ResponseWriter, r *http.Request) {
	anims, err := GetAnims()
	CheckHttpErr(err, w, r, 500)
	var d template.HTMLDataMap
	err = DB.Get("pages", "anims", &d)
	if CheckHttpErr(err, w, r, 500) {
		return
	}
	d.Set("Anims", anims)
	d.Set("NoAnims", len(anims))
	err = template.Use(w, r, "anims", d)
	CheckHttpErr(err, w, r, 500)
}

func AnimManageList(w http.ResponseWriter, r *http.Request) {
	anims, err := GetAnims()
	if CheckHttpErr(err, w, r, 500) {
		return
	}
	opts := make([]UrlOpt, len(anims)+1)
	opts[0] = UrlOpt{Name: "Add new animation", URL: "/manage/animation/?id=new"}
	for i, a := range anims {
		opts[i+1] = UrlOpt{Name: a.Title, URL: "/manage/animation/?id=" + a.ID}
	}
	d := template.Data("Manage animations", "List of animations")
	d.Set("Options", opts)
	err = template.Use(w, r, "manage", d)
	CheckHttpErr(err, w, r, 500)
	return
}

func AnimManager(w http.ResponseWriter, r *http.Request) {
	if "/manage/animation/" != r.URL.Path {
		HttpErr(w, r, 404)
		return
	}
	if r.Method == "GET" {
		id := r.URL.Query().Get("id")
		if id == "" {
			AnimManageList(w, r)
			return
		}
		var a AnimInfo
		var err error
		if id == "new" {
			a.ID = "new"
		} else {
			a, err = GetAnim(id)
		}
		if CheckHttpErr(err, w, r, 404) {
			return
		}
		data := template.Data("Edit animation", id)
		data.Set("AnimUrl", "/manage/animation/")
		data.Set("Anim", a)
		err = template.Use(w, r, "edit", data)
		CheckHttpErr(err, w, r, 500)
		return
	}
	if r.Method == "POST" {
		err := r.ParseForm()
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		g := func(s string) string {
			s = r.Form.Get(s)
			return strings.TrimSpace(s)
		}
		gg := func(s string) []string {
			var ss []string
			for _, s = range strings.Split(g(s), "\n") {
				s = strings.TrimSpace(s)
				if s != "" {
					ss = append(ss, s)
				}
			}
			return ss
		}
		var a AnimInfo
		a.ID = g("id")
		a.Title = g("title")
		a.Date = g("date")
		a.Desc = g("desc")
		a.Imgs = gg("imgs")
		a.Vids = gg("vids")
		if a.ID == "new" {
			a = GenAnimID(a)
		}
		if !a.HasReqFields() {
			HttpErr(w, r, 400)
			return
		}
		err = DB.Save(&a)
		if CheckHttpErr(err, w, r, 400) {
			return
		}
	}
	http.Redirect(w, r, "/animations/", http.StatusFound)
}
