package main

import (
	"go.etcd.io/bbolt"
	"mime"
	"net/http"
	"path/filepath"
	"sophuwu.site/myweb/template"
	"strings"
)

type DBFile struct {
	Name string
	Size int
}

func (f *DBFile) Valid() bool {
	if f.Name == "" || strings.HasPrefix(f.Name, ".") || strings.HasPrefix(f.Name, "_") || f.Size == 0 {
		return false
	}
	return true
}
func (f *DBFile) Set(name string, size int) {
	f.Name = name
	f.Size = size
}

func MediaHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/media/")
	var err error
	if path == "" {
		var list []DBFile
		var t DBFile
		err = DB.Bolt.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("media"))
			return b.ForEach(func(k, v []byte) error {
				t.Set(string(k), len(v))
				if t.Valid() {
					list = append(list, t)
				}
				return nil
			})
		})
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		d := template.Data("/media/", "Directory listing for /media/")
		d.Set("Files", list)
		d.Set("NoFiles", len(list))
		err = template.Use(w, r, "filelist", d)
		CheckHttpErr(err, w, r, 500)
		return
	}
	var data []byte
	data, err = DB.GetBytes("media", path)
	if CheckHttpErr(err, w, r, 404) {
		return
	}
	w.WriteHeader(200)
	w.Header().Set("content-type", mime.TypeByExtension(filepath.Ext(path)))
	w.Write(data)
}
