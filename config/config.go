package config

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
)

var (
	ListenAddr string
	WebRoot    string
	DbPath     string
	StaticPath string
	MediaPath  string
	Templates  string
	Email      string
	Name       string
	URL        string
)

func path(p string) string {
	return filepath.Join(WebRoot, p)
}

func init() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <config file>", os.Args[0])
	}
	file, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Error opening config: %v", err)
	}

	var mm = make(map[string]string)
	var pos int

	for _, v := range bytes.Split(file, []byte{'\n'}) {
		if len(v) == 0 || (len(v) > 0 && v[0] == '#') || func() bool {
			pos = bytes.IndexByte(v, '=')
			return pos == -1
		}() {
			continue
		}
		mm[string(bytes.TrimSpace(v[:pos]))] = string(bytes.TrimSpace(v[pos+1:]))
	}
	ListenAddr = mm["address"] + ":" + mm["port"]
	if ListenAddr[len(ListenAddr)-1] == ':' {
		ListenAddr += "8085"
	}

	if len(mm["webroot"]) > 0 && mm["webroot"][0] == '/' {
		WebRoot = mm["webroot"]
	} else {
		WebRoot = filepath.Join(filepath.Dir(os.Args[1]), mm["webroot"])
	}
	DbPath = path("stuff.sqlite")
	StaticPath = path("static")
	MediaPath = path("media")
	Templates = path("templates/*")
	Email = mm["email"]
	Name = mm["name"]
	URL = mm["url"]

}
