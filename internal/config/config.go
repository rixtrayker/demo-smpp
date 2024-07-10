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
	SSLMode  string `koanf:"sslmode"`
	MaxConn  int    `koanf:"max_conn"`
	MaxIdle  int    `koanf:"max_idle"`
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
	RateLimit      int      `koanf:"rate_limit"`
	BurstLimit     int      `koanf:"burst_limit"`
	MaxOutStanding int      `koanf:"max_outstanding"`
	HasOutStanding bool     `koanf:"has_outstanding"`
	MaxRetries     int      `koanf:"max_retries"`
	Queues         []string `koanf:"queues"`
}

var config *Config
var k = koanf.New(".")

func LoadConfig(filePath string) *Config {
	if config != nil {
		return config
	}

	if filePath == "" {
		filePath = "config.yaml"
	}

	if err := k.Load(file.Provider(filePath), yaml.Parser()); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	config = &Config{}
	if err := k.Unmarshal("", config); err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	setDefaults(config)

	// Derive SMSC field in providers
	for i := range config.ProvidersConfig {
		config.ProvidersConfig[i].SMSC = fmt.Sprintf("%s:%d", config.ProvidersConfig[i].Address, config.ProvidersConfig[i].Port)
		logProviderConfig(config.ProvidersConfig[i])
	}

	return config
}

func setDefaults(cfg *Config) {
	if cfg.RateLimit == 0 {
		cfg.RateLimit = 1000
	}

	for i := range cfg.ProvidersConfig {
		p := &cfg.ProvidersConfig[i]
		if p.RateLimit == 0 {
			p.RateLimit = 100
		}
		if p.BurstLimit == 0 {
			p.BurstLimit = 10
		}
		if p.MaxOutStanding == 0 {
			p.MaxOutStanding = 1000
		}
		if !p.HasOutStanding {
			p.HasOutStanding = true
		}
		if p.MaxRetries == 0 {
			p.MaxRetries = 3
		}
		if p.Queues == nil {
			p.Queues = []string{"default"}
		}
	}

	if cfg.DatabaseConfig.Port == 0 {
		cfg.DatabaseConfig.Port = 5432 // default port for PostgreSQL
	}
}

func logProviderConfig(provider Provider) {
	logrus.WithFields(logrus.Fields{
		"Name":           provider.Name,
		"SessionType":    provider.SessionType,
		"SMSC":           provider.SMSC,
		"SystemID":       provider.SystemID,
		"Password":       provider.Password,
		"SystemType":     provider.SystemType,
		"RateLimit":      provider.RateLimit,
		"BurstLimit":     provider.BurstLimit,
		"MaxOutStanding": provider.MaxOutStanding,
		"HasOutStanding": provider.HasOutStanding,
		"MaxRetries":     provider.MaxRetries,
		"Queues":         provider.Queues,
	}).Info("Provider")
}
