package main

import (
	"context"
	"errors"
	"fmt"
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

var HttpErrs = map[int]string{
	401: "Unauthorized: You must log in to access this page.",
	403: "Forbidden: You do not have permission to access this page.",
	404: "Not found: the requested page does not exist. Please check the URL and try again.",
	500: "Internal server error: the server encountered an error while processing your request. Please try again later.",
}

func HttpErr(w http.ResponseWriter, r *http.Request, code int) {
	w.WriteHeader(code)
	var ErrTxt string
	if t, ok := HttpErrs[code]; ok {
		ErrTxt = t
	} else {
		ErrTxt = "An error occurred. Please try again later."
	}
	data := template.Data("An error occurred", fmt.Sprintf("%d: %s", code, ErrTxt))
	data.Set("ErrText", ErrTxt)
	data.Set("ErrCode", code)
	err := template.Use(w, r, "err", data)
	if err != nil {
		log.Printf("error writing error page: %v", err)
	}
}

func HttpIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		HttpErr(w, r, 404)
		return
	}
	d, err := GetPageData("index")
	if CheckHttpErr(err, w, r, 500) {
		return
	}
	err = template.Use(w, r, "index", d)
	_ = CheckHttpErr(err, w, r, 500)
}

func HttpFS(path, fspath string) {
	http.Handle(path, http.StripPrefix(path, http.FileServer(http.Dir(fspath))))
}

func main() {
	OpenDB()
	err := template.Init(config.Templates)
	if err != nil {
		log.Fatalf("Error initializing templates: %v", err)
	}

	http.HandleFunc("/", HttpIndex)
	http.HandleFunc("/blog/", BlogHandler)
	HttpFS("/static/", config.StaticPath)
	http.HandleFunc("/media/", MediaHandler)
	// HttpFS("/media/", config.MediaPath)

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
