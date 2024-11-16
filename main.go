package main

import (
	"html/template"
	"log"
	"net/http"
	"sophuwu.site/myweb/config"
)

var Tplt *template.Template

func ParseTemplates() {
	Tplt = template.Must(template.ParseGlob(config.Templates))
}

func HttpIndex(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]string)
	data["Title"] = config.Name
	data["Description"] = "Blogs and projects by " + config.Name + "."
	data["Url"] = config.URL + r.URL.Path
	data["Email"] = config.Email
	data["Name"] = config.Name

	if err := Tplt.ExecuteTemplate(w, "index", data); err != nil {
		log.Println(err)
		return
	}
}

func HttpFS(path, fspath string) {
	http.Handle(path, http.StripPrefix(path, http.FileServer(http.Dir(fspath))))
}

func main() {
	ParseTemplates()

	http.HandleFunc("/", HttpIndex)
	HttpFS("/static/", config.StaticPath)
	HttpFS("/media/", config.MediaPath)

	http.ListenAndServe(config.ListenAddr, nil)

}
