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
	Logging    Logging
	Database   Database
	HCaptcha   HCaptcha
}

type Auth struct {
	TLSCert          string `yaml:"tls_cert"`
	TLSKey           string `yaml:"tls_key"`
	CookieMaxAgeDays int    `yaml:"cookie_max_age_days"`
	CookieSecret     string `yaml:"cookie_secret"`
}

type HCaptcha struct {
	SecretKey string `yaml:"secret_key"`
	SiteKey   string `yaml:"site_key"`
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

type Logging struct {
	LogLevel string `yaml:"log_level"`
	LogFile  string `yaml:"log_file"`
}
