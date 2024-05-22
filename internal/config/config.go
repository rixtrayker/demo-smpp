package config

import (
	"os"
	"strconv"
	"strings"
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
    MaxOutStanding int
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
    maxOutstanding, err := strconv.Atoi(strings.TrimSpace(os.Getenv(provider + "_MAX_OUTSTANDING")))
    if err != nil {
        maxOutstanding = 100
    }
    return Provider{
        Name:           strings.TrimSpace(os.Getenv(provider + "_NAME")),
        SMSC:           strings.TrimSpace(os.Getenv(provider + "_SMSC")),
        SystemID:       strings.TrimSpace(os.Getenv(provider + "_SYSTEM_ID")),
        Password:       strings.TrimSpace(os.Getenv(provider + "_PASSWORD")),
        SystemType:     strings.TrimSpace(os.Getenv(provider + "_SYSTEM_TYPE")),
        MaxOutStanding: maxOutstanding,
    }
}