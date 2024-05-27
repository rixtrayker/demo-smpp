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
    MaxRetries int
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
    host := os.Getenv("SMPP_HOST")
    port := os.Getenv("SMPP_PORT")
    return SMPPConfig{
        SMSC:       host + ":" + port,
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
    host := strings.TrimSpace(os.Getenv(provider + "_HOST"))
    port := strings.TrimSpace(os.Getenv(provider + "_PORT"))
    smsc := host + ":" + port
    name := strings.TrimSpace(os.Getenv(provider + "_NAME"))
    systemID := strings.TrimSpace(os.Getenv(provider + "_SYSTEM_ID"))
    password := strings.TrimSpace(os.Getenv(provider + "_PASSWORD"))
    systemType := strings.TrimSpace(os.Getenv(provider + "_SYSTEM_TYPE"))

    if name == "" || smsc == "" || systemID == "" || password == "" {
        return Provider{}
    }

    maxOutstanding, err := strconv.Atoi(strings.TrimSpace(os.Getenv(provider + "_MAX_OUTSTANDING")))
    globalMaxOutstanding := 100    
    if err != nil {
        maxOutstanding = globalMaxOutstanding
    }

    maxRetries, err := strconv.Atoi(strings.TrimSpace(os.Getenv(provider + "_MAX_RETRIES")))
    globalMaxRetries := 3
    if err != nil {
        maxRetries = globalMaxRetries
    }

    return Provider{
        Name:           name,
        SMSC:           smsc,
        SystemID:       systemID,
        Password:       password,
        SystemType:     systemType,
        MaxOutStanding: maxOutstanding,
        MaxRetries:     maxRetries,
    }
}