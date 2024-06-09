package config

import (
	"fmt"
	"log"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/sirupsen/logrus"
)

type Config struct {
	RedisURL        string        `koanf:"redis_url"`
	RateLimit       int           `koanf:"rate_limit"`
	SMPPConfig      SMPPConfig    `koanf:"smpp_config"`
	DatabaseConfig  DatabaseConfig `koanf:"database_config"`
	ProvidersConfig []Provider    `koanf:"providers"`
}

type SMPPConfig struct {
	SMSC       string `koanf:"smsc"`
	SystemID   string `koanf:"system_id"`
	Password   string `koanf:"password"`
	SystemType string `koanf:"system_type"`
}

type DatabaseConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	DBName   string `koanf:"dbname"`
}

type Provider struct {
	Name           string   `koanf:"name"`
	SessionType    string   `koanf:"session_type"`
	SMSC           string   `koanf:"smc"` // This will be derived from Address and Port
	Address        string   `koanf:"address"`
	Port           int      `koanf:"port"`
	SystemID       string   `koanf:"system_id"`
	Password       string   `koanf:"password"`
	SystemType     string   `koanf:"system_type"`
	MaxOutStanding int      `koanf:"max_outstanding"`
	HasOutStanding bool     `koanf:"has_outstanding"`
	MaxRetries     int      `koanf:"max_retries"`
	Queues         []string `koanf:"queues"`
}

var config *Config
var k = koanf.New(".")

func LoadConfig() *Config {
	if config != nil {
		return config
	}

	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	config = &Config{}
	if err := k.Unmarshal("", config); err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	// Derive SMSC field in providers
	for i := range config.ProvidersConfig {
		config.ProvidersConfig[i].SMSC = fmt.Sprintf("%s:%d", config.ProvidersConfig[i].Address, config.ProvidersConfig[i].Port)
		logProviderConfig(config.ProvidersConfig[i])
	}

	return config
}

func logProviderConfig(provider Provider) {
	logrus.WithFields(logrus.Fields{
		"Name":           provider.Name,
		"SessionType":    provider.SessionType,
		"SMSC":           provider.SMSC,
		"SystemID":       provider.SystemID,
		"Password":       provider.Password,
		"SystemType":     provider.SystemType,
		"MaxOutStanding": provider.MaxOutStanding,
	}).Info("Provider")
}
