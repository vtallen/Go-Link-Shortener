package conf

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

type Config struct {
	Shortcodes Shortcodes
	Auth       Auth
	Server     Server
	Database   Database
}

type Auth struct {
	ApiKeyLen        int    `yaml:"api_key_length"`
	RootUsername     string `yaml:"root_username"`
	RootPassword     string `yaml:"root_password"`
	TLSCert          string `yaml:"tls_cert"`
	TLSKey           string `yaml:"tls_key"`
	CookieMaxAgeDays int    `yaml:"cookie_max_age_days"`
	CookieSecret     string `yaml:"cookie_secret"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Database struct {
	Path string `yaml:"path"`
}

type Shortcodes struct {
	ShortcodeLength int    `yaml:"shortcode_length"`
	Universe        string `yaml:"shortcode_universe"`
}
