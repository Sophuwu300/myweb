package main

import (
	"context"
	"errors"
	"golang.org/x/sys/unix"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sophuwu.site/myweb/config"
	"sophuwu.site/myweb/template"
)

func CheckHttpErr(err error, w http.ResponseWriter, r *http.Request, code int) bool {
	if err != nil {
		HttpErr(w, r, code)
		log.Printf("err: %v: HTTP %d: %s %s\n", err, code, r.Method, r.URL.Path)
		return true
	}
	return false
}

func HttpErr(w http.ResponseWriter, r *http.Request, code int) {
	http.Error(w, http.StatusText(code), code)
}

func HttpIndex(w http.ResponseWriter, r *http.Request) {
	d, err := GetPageData("index")
	if CheckHttpErr(err, w, r, 500) {
		return
	}
	err = template.Use(w, r, "index", d)
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
	OpenDB()
	err := template.Init(config.Templates)
	if err != nil {
		log.Fatalf("Error initializing templates: %v", err)
	}

	http.HandleFunc("/", HttpIndex)
	http.HandleFunc("/blog/", BlogHandler)
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
