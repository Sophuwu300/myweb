package config

import (
	yaml "gopkg.in/yaml.v3"
	"log"
	"os"
)

// Server is a struct that holds the configuration for the server
type Server struct {
	Port string `yaml:"Port"`
	IP   string `yaml:"IP"`
}

// Host is function that returns the host string
func (s Server) Host() string {
	return s.IP + ":" + s.Port
}

// Dirs is a struct that holds directories needed for the server
type Dirs struct {
	Static    string `yaml:"Static"`
	Templates string `yaml:"Templates"`
	Media     string `yaml:"Media"`
}

// Link is a struct that holds the configuration for a link
type Link struct {
	Name string `yaml:"Name"`
	URL  string `yaml:"URL"`
}

// ContactInfo is a struct that holds the configuration for the contact information
type ContactInfo struct {
	Email string `yaml:"Email"`
	Name  string `yaml:"Name"`
	Links []Link `yaml:"Links"`
}

// WebInfo is a struct that holds the configuration for the website information
type WebInfo struct {
	Title       string `yaml:"Title"`
	Description string `yaml:"Description"`
	Url         string `yaml:"Url"`
}

// Config is a struct that holds the configuration for the server, directories, and contact information
type Config struct {
	Server  Server      `yaml:"Server"`
	Paths   Dirs        `yaml:"Paths"`
	Contact ContactInfo `yaml:"Contact"`
	Website WebInfo     `yaml:"Website"`
}

// Cfg is a variable that holds the configuration
var Cfg Config

func init() {
	var path string = "config.yaml"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error opening config: %v", err)
	}
	err = yaml.Unmarshal(file, &Cfg)
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}
}
