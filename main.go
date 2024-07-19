package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sophuwu.site/myweb/config"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		tpath := filepath.Join(config.Cfg.Paths.Templates, "*")
		t := template.Must(template.ParseGlob(tpath))
		data := make(map[string]string)
		data["Title"] = config.Cfg.Website.Title
		data["WebsiteTitle"] = config.Cfg.Website.Title
		data["Description"] = config.Cfg.Website.Description
		data["Url"] = config.Cfg.Website.Url + r.URL.Path

		if err := t.ExecuteTemplate(w, "index", data); err != nil {
			log.Println(err)
			return
		}
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(config.Cfg.Paths.Static))))
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(config.Cfg.Paths.Media))))

	http.ListenAndServe(config.Cfg.Server.Host(), nil)

}
