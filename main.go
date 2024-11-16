package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sophuwu.site/myweb/config"
)

var Tplt *template.Template

func ParseTemplates() {
	Tplt = template.Must(template.ParseGlob(filepath.Join(config.Cfg.Paths.Templates, "*")))
}

func HttpIndex(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]string)
	data["Title"] = config.Name
	data["Description"] = "Blogs and projects by " + config.Name + "."
	data["Url"] = config.Cfg.Website.Url + r.URL.Path
	data["Email"] = config.Cfg.Contact.Email
	data["Name"] = config.Cfg.Contact.Name

	if err := Tplt.ExecuteTemplate(w, "index", data); err != nil {
		log.Println(err)
		return
	}
}

func main() {
	ParseTemplates()

	http.HandleFunc("/", HttpIndex)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(config.Cfg.Paths.Static))))
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(config.Cfg.Paths.Media))))

	http.ListenAndServe(config.ListenAddr(), nil)

}
