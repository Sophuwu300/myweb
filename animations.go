package main

import (
	"crypto/md5"
	"encoding/base64"
	"git.sophuwu.com/myweb/template"
	"net/http"
	"strings"
	"time"
)

// AnimInfo is a struct that holds information about an animation.
type AnimInfo struct {
	ID    string `storm:"id"`
	Title string
	Date  string `storm:"index"`
	Desc  string
	Imgs  []string
	Vids  []string
}

// HasReqFields checks if all required fields have a non-empty value.
func (a *AnimInfo) HasReqFields() bool {
	return a.Title != "" && a.Desc != "" && (len(a.Imgs)+len(a.Vids) > 0) && a.Date != "" && a.ID != ""
}

// GenAnimID generates an ID for an animation. It will generate a different ID
// each time it is called, even if the input is the same. This allows for
// multiple animations with the same title to be stored without conflict.
func GenAnimID(a AnimInfo) AnimInfo {
	md := md5.New()
	md.Write([]byte(time.Now().String() + a.Title))
	a.ID = base64.URLEncoding.EncodeToString(md.Sum(nil))
	if a.Date == "" {
		a.Date = time.Now().Format("2006-01-02")
	}
	return a
}

// GetAnim retrieves AnimInfo from the database with the given ID.
// If the ID is not found, an error is returned.
func GetAnim(id string) (AnimInfo, error) {
	var a AnimInfo
	err := DB.One("ID", id, &a)
	return a, err
}

// AnimDelete deletes an animation from the database with the given ID.
func AnimDelete(id string) error {
	return DB.DeleteStruct(&AnimInfo{ID: id})
}

// GetAnims retrieves all animations from the database. The animations are
// sorted by date, with the most recent first in []AnimInfo.
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

// AnimHandler is a http.HandlerFunc that serves the animations page.
// It retrieves all animations from the database and displays them.
func AnimHandler(w http.ResponseWriter, r *http.Request) {
	anims, err := GetAnims()
	CheckHttpErr(err, w, r, 500)
	d, err := GetPageData("anims")
	if CheckHttpErr(err, w, r, 500) {
		return
	}
	d.Set("Anims", anims)
	d.Set("NoAnims", len(anims))
	err = template.Use(w, r, "anims", d)
	CheckHttpErr(err, w, r, 500)
}

// AnimManageList is a http.HandlerFunc that serves the animation manager list.
// It retrieves all animations from the database and displays them as a list.
// With each animation, there is a link to edit the details of that animation.
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

// AnimManager is a http.HandlerFunc that serves the animation manager. It
// allows the user to edit an existing animation or create a new one.
// If the ID is "new", a new animation is created. Otherwise, the animation
// with the given ID is retrieved from the database and displayed for editing.
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
