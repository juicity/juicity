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
	Server        string `json:"server"`
	Uuid          string `json:"uuid"`
	Password      string `json:"password"`
	Sni           string `json:"sni"`
	AllowInsecure bool   `json:"allow_insecure"`

	// Server
	Users       map[string]string `json:"users"`
	Certificate string            `json:"certificate"`
	PrivateKey  string            `json:"private_key"`
	SendThrough string            `json:"send_through"`

	// Common
	Listen            string `json:"listen"`
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
