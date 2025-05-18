package template

import (
	"bytes"
	"git.sophuwu.com/myweb/config"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

type HTMLDataMap map[string]any

func Data(title, desc string) HTMLDataMap {
	var data = make(HTMLDataMap)
	data["Title"] = title
	data["Description"] = desc
	return data
}

func (d *HTMLDataMap) Set(key string, value any) {
	(*d)[key] = value
}

func (d *HTMLDataMap) SetHTML(value string) {
	(*d)["Content"] = value
}

func (d *HTMLDataMap) SetIfEmpty(key string, value any) {
	if _, ok := (*d)[key]; !ok {
		(*d)[key] = value
	}
}

var templates *template.Template
var fillFunc func(w http.ResponseWriter, name string, data HTMLDataMap) error
var templatesDir string

func FillString(name string, data HTMLDataMap) (string, error) {
	data.SetIfEmpty("Url", config.URL)
	data.SetIfEmpty("Email", config.Email)
	data.SetIfEmpty("Name", config.Name)
	if data["Content"] != nil {
		data["HTML"] = template.HTML(data["Content"].(string))
	}
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func parseTemplates() error {
	index := template.New("index")
	index.Parse(filepath.Join(templatesDir, "index.html"))
	index.Option()

	tmp, err := template.ParseGlob(templatesDir)
	if err != nil {
		return err
	}
	templates = tmp
	return nil
}

func Init(path string) error {
	templatesDir = path
	if os.Getenv("DEBUG") == "1" {
		fillFunc = func(w http.ResponseWriter, name string, data HTMLDataMap) error {
			err := parseTemplates()
			if err != nil {
				return err
			}
			return templates.ExecuteTemplate(w, name, data)
		}
	} else {
		fillFunc = func(w http.ResponseWriter, name string, data HTMLDataMap) error {
			return templates.ExecuteTemplate(w, name, data)
		}
	}

	return parseTemplates()
}

func Use(w http.ResponseWriter, r *http.Request, name string, data HTMLDataMap) error {
	data.SetIfEmpty("Url", config.URL+r.URL.Path)
	data.SetIfEmpty("Email", config.Email)
	data.SetIfEmpty("Name", config.Name)
	if data["Content"] != nil {
		data["HTML"] = template.HTML(data["Content"].(string))
	}
	return fillFunc(w, name, data)
}
