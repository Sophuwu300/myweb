package config

import (
	yaml "gopkg.in/yaml.v3"
	"log"
	"os"
)

var (
	ListenAddr  string
	TemplateDir string
	StaticDir   string
	MediaDir    string
	DBPath      string
	Email       string
	Name        string
	URL         string
)

var m map[string]string

func init() {
	var path string = "config.yaml"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error opening config: %v", err)
	}
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}
}
