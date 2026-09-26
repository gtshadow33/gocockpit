package config

import "github.com/BurntSushi/toml"

type Config struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	AdminGroup string `toml:"admin_group"`
}

var App Config

func Load(path string) error {
	_, err := toml.DecodeFile(path, &App)
	return err
}