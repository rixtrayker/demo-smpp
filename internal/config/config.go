package config

import (
	"os"
)

type Config struct {
    RedisURL         string
    RateLimit        int
    SMPPConfig       SMPPConfig
    DatabaseConfig   DatabaseConfig
    ProvidersConfig []Provider
}

type SMPPConfig struct {
    SMSC       string
    SystemID   string
    Password   string
    SystemType string
}

type DatabaseConfig struct {
    // Add necessary fields
    DSN string
}

type Provider struct {
    Name     string
    SMSC       string
    SystemID   string
    Password   string
    SystemType string
    // Host     string
    // Port     int
    // Username string
    // Password string
}

func LoadConfig() Config {
    return Config{
        RedisURL:         os.Getenv("REDIS_URL"),
        RateLimit:        100, // For example
        SMPPConfig:       loadSMPPConfig(),
        DatabaseConfig:   loadDatabaseConfig(),
        ProvidersConfig:   loadProvidersConfig(),
    }
}

func loadSMPPConfig() SMPPConfig {
    return SMPPConfig{
        SMSC:       os.Getenv("SMSC"),
        SystemID:   os.Getenv("SYSTEM_ID"),
        Password:   os.Getenv("PASSWORD"),
        SystemType: os.Getenv("SYSTEM_TYPE"),
    }
}

func loadDatabaseConfig() DatabaseConfig {
    return DatabaseConfig{
        DSN: os.Getenv("DATABASE_URL"),
    }
}

func loadProvidersConfig() []Provider {
    return []Provider{
        loadProviderConfig("PROVIDER_A"),
        loadProviderConfig("PROVIDER_B"),
        loadProviderConfig("PROVIDER_C"),
    }
}

func loadProviderConfig(provider string) Provider {
    return Provider{
        Name:     os.Getenv(provider + "_NAME"),
        SMSC:       os.Getenv(provider + "_SMSC"),
        SystemID:   os.Getenv(provider + "_SYSTEM_ID"),
        Password:   os.Getenv(provider + "_PASSWORD"),
    }
}