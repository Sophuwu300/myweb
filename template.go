package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sophuwu.site/myweb/config"
)

type HTMLTemplates_t struct {
	templates *template.Template
	fillFunc  func(w http.ResponseWriter, name string, data HTMLDataMap) error
}

type HTMLDataMap map[string]any

func HTMLData(title, desc string) HTMLDataMap {
	var data = make(HTMLDataMap)
	data["Title"] = title
	data["Description"] = desc
	return data
}

func (d *HTMLDataMap) Set(key string, value any) {
	(*d)[key] = value
}

func (d *HTMLDataMap) SetHTML(key string, value string) {
	(*d)[key] = template.HTML(value)
}

func (d *HTMLDataMap) SetIfEmpty(key string, value any) {
	if _, ok := (*d)[key]; !ok {
		(*d)[key] = value
	}
}

var Templates HTMLTemplates_t

func (h *HTMLTemplates_t) ParseTemplates() error {
	index := template.New("index")
	index.Parse(filepath.Join(config.Templates, "index.html"))
	index.Option()

	tmp, err := template.ParseGlob(config.Templates)
	if err != nil {
		return err
	}
	h.templates = tmp
	return nil
}

func (h *HTMLTemplates_t) Init() error {
	if os.Getenv("DEBUG") == "1" {
		h.fillFunc = func(w http.ResponseWriter, name string, data HTMLDataMap) error {
			err := h.ParseTemplates()
			if err != nil {
				return err
			}
			return h.templates.ExecuteTemplate(w, name, data)
		}
	} else {
		h.fillFunc = func(w http.ResponseWriter, name string, data HTMLDataMap) error {
			return h.templates.ExecuteTemplate(w, name, data)
		}
	}

	return h.ParseTemplates()
}

func (h *HTMLTemplates_t) Use(w http.ResponseWriter, r *http.Request, name string, data HTMLDataMap) error {
	data.SetIfEmpty("Url", config.URL+r.URL.Path)
	data.SetIfEmpty("Email", config.Email)
	data.SetIfEmpty("Name", config.Name)
	if data["Content"] != nil {
		data["HTML"] = template.HTML(data["Content"].(string))
	}
	return h.fillFunc(w, name, data)
}
