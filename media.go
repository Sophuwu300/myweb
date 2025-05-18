package main

import (
	"bytes"
	"fmt"
	"git.sophuwu.com/myweb/template"
	"go.etcd.io/bbolt"
	"io"
	"mime"
	"net/http"
	"os/exec"
	"path/filepath"
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

func ListMedia() ([]DBFile, error) {
	var list []DBFile
	var t DBFile
	err := DB.Bolt.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("media"))
		return b.ForEach(func(k, v []byte) error {
			t.Set(string(k), len(v))
			if t.Valid() {
				list = append(list, t)
			}
			return nil
		})
	})
	return list, err
}

func MediaHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/media/")
	var err error
	if path == "" {
		var list []DBFile
		list, err = ListMedia()
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

func AddMedia(path string, data []byte) error {
	return DB.SetBytes("media", path, data)
}

func ConvWebp(f io.Reader) (bytes.Buffer, error) {
	cmd := exec.Command("convert", "-", "webp:-")
	var data bytes.Buffer
	cmd.Stdin = f
	cmd.Stdout = &data
	err := cmd.Run()
	return data, err
}

func ManageMedia(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/manage/media/" {
		HttpErr(w, r, 404)
		return
	}
	if r.Method == "GET" {
		d := template.Data("Manage media", "Upload media files")
		d.Set("Media", "true")
		err := template.Use(w, r, "edit", d)
		CheckHttpErr(err, w, r, 500)
		return
	}
	if r.Method == "POST" {

		err := r.ParseMultipartForm(10 << 20)
		if CheckHttpErr(err, w, r, 500) {
			return
		}

		fh, h, err := r.FormFile("file1")
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		defer fh.Close()
		f := io.Reader(fh)
		ext := filepath.Ext(h.Filename)
		ext = strings.TrimPrefix(ext, ".")
		name := filepath.Base(h.Filename)
		var data bytes.Buffer
		if ext == "jpg" || ext == "jpeg" || ext == "png" || ext == "gif" {
			// convert to webp
			name = strings.TrimSuffix(name, ext)
			if !strings.HasSuffix(name, ".webp") {
				name += ".webp"
			}
			name = strings.ReplaceAll(name, " ", "-")
			name = strings.ReplaceAll(name, "..", ".")
			data, err = ConvWebp(f)
		} else {
			_, err = data.ReadFrom(f)
		}
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		// add to db
		err = AddMedia(name, data.Bytes())
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		http.Redirect(w, r, "/media/", 302)
	}
}

func DeleteMedia(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/manage/delete/media/")
	if path == "" {
		list, err := ListMedia()
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		d := template.Data("Delete media", "Delete media files")
		d.Set("Files", list)
		d.Set("NoFiles", len(list))
		err = template.Use(w, r, "filelist", d)
		CheckHttpErr(err, w, r, 500)
		return
	}
	conf := r.URL.Query().Get("confirm")
	if conf == "true" {
		err := DB.Delete("media", path)
		if CheckHttpErr(err, w, r, 500) {
			return
		}
		http.Redirect(w, r, "/media/", 302)
	}
	w.Header().Set("content-type", "text/html")
	w.WriteHeader(200)
	fmt.Fprintf(w, "Are you sure you want to delete %s?<br><a href=\"/manage/delete/media/%s?confirm=true\">Yes</a><br><a href=\"/\">No</a>", path, path)
}
