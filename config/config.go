package config

import (
	"encoding/json"
	"os"
)

var (
	Version = "unknown"
)

type Config struct {
	// Client
	Server   string `json:"server"`
	Uuid     string `json:"uuid"`
	Password string `json:"password"`
	Sni      string `json:"sni"`

	// Server
	Listen      string            `json:"listen"`
	Users       map[string]string `json:"users"`
	Certificate string            `json:"certificate"`
	PrivateKey  string            `json:"private_key"`

	// Common
	CongestionControl string `json:"congestion_control"`
	LogLevel          string `json:"log_level"`
}

func ReadConfig(p string) (*Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var c Config
	if err = json.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
