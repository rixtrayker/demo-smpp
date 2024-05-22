package config

import (
    "os"
)

type Config struct {
    RedisURL         string
    RateLimit        int
    SMPPConfig       SMPPConfig
    DatabaseConfig   DatabaseConfig
}

type SMPPConfig struct {
    // Add necessary fields
    SMSC       string
    SystemID   string
    Password   string
    SystemType string
}

type DatabaseConfig struct {
    // Add necessary fields
    DSN string
}

func LoadConfig() Config {
    return Config{
        RedisURL:         os.Getenv("REDIS_URL"),
        RateLimit:        100, // For example
        SMPPConfig:       loadSMPPConfig(),
        DatabaseConfig:   loadDatabaseConfig(),
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
