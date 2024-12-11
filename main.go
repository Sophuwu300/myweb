package main

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sophuwu.site/myweb/config"
)

func Sha1Base64(data ...any) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprint(data...)))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func CheckHttpErr(err error, w http.ResponseWriter, r *http.Request, code int) bool {
	if err != nil {
		HttpErr(w, r, code)
		return true
	}
	return false
}

func HttpErr(w http.ResponseWriter, r *http.Request, code int) {
	http.Error(w, http.StatusText(code), code)
	log.Printf("HTTP %d: %s %s\n", code, r.Method, r.URL.Path)
}

func HttpIndex(w http.ResponseWriter, r *http.Request) {
	var d HTMLDataMap
	err := DB.Get("pages", "index", &d)
	if CheckHttpErr(err, w, r, 500) {
		return
	}
	err = Templates.Use(w, r, "index", d)
	_ = CheckHttpErr(err, w, r, 500)
}

type HttpHjk struct {
	http.ResponseWriter
	status int
}

func HttpFS(path, fspath string) (string, http.HandlerFunc) {
	fileServer := http.StripPrefix(path, http.FileServer(http.Dir(fspath)))
	return path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijack := &HttpHjk{ResponseWriter: w}
		fileServer.ServeHTTP(hijack, r)
		if hijack.status >= 400 && hijack.status < 600 {
			HttpErr(w, r, hijack.status)
		}
	})
}

func main() {
	err := Templates.Init()
	OpenDB()

	d := HTMLData(config.Name, fmt.Sprintf("About %s. look at animations I've made, read about things I've found interesting. Links to my social media.", config.Name))
	d.SetHTML("Content", "<h1>Welcome to my website</h1><p>Here you can find animations I've made, blogs I've written, and other things I've found interesting.</p>")
	DB.Set("pages", "index", &d)

	if err != nil {
		log.Fatalf("Error initializing templates: %v", err)
	}

	http.HandleFunc("/", HttpIndex)
	http.HandleFunc(HttpFS("/static/", config.StaticPath))
	http.HandleFunc(HttpFS("/media/", config.MediaPath))

	server := http.Server{Addr: config.ListenAddr, Handler: nil}
	go func() {
		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error starting server: %v", err)
		}
	}()
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, unix.SIGINT, unix.SIGTERM, unix.SIGQUIT, unix.SIGKILL, unix.SIGSTOP)
	s := <-sigchan
	println("stopping: got signal", s.String())
	err = server.Shutdown(context.Background())
	if err != nil {
		log.Println("Error stopping server: %v", err)
	}
	CloseDB()
	println("stopped")
}
