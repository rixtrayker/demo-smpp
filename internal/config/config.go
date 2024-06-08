package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	RedisURL        string
	RateLimit       int
	SMPPConfig      SMPPConfig
	DatabaseConfig  DatabaseConfig
	ProvidersConfig []Provider
}

type SMPPConfig struct {
	SMSC       string
	SystemID   string
	Password   string
	SystemType string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type Provider struct {
	Name           string
	SessionType    string
	SMSC           string
	Address 	   string 
	Port 		   int
	SystemID       string
	Password       string
	SystemType     string
	MaxOutStanding int
	HasOutStanding bool
	MaxRetries     int
	Queues         []string
}

var config *Config

func LoadConfig() *Config {
	if config != nil {
		return config
	}

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // required if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	viper.AutomaticEnv()          // read in environment variables that match

	// Set default values
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file, %s", err)
	}

	config = &Config{
		RedisURL:        viper.GetString("redis_url"),
		RateLimit:       viper.GetInt("rate_limit"),
		SMPPConfig:      loadSMPPConfig(),
		DatabaseConfig:  loadDatabaseConfig(),
		ProvidersConfig: loadProvidersConfig(),
	}

	return config
}

func setDefaults() {
	viper.SetDefault("rate_limit", 100)
	viper.SetDefault("database_config.port", 3306)
	viper.SetDefault("database_config.host", "localhost")
	viper.SetDefault("database_config.user", "root")
	viper.SetDefault("database_config.password", "")
	viper.SetDefault("database_config.dbname", "go_client")
}

func loadSMPPConfig() SMPPConfig {
	return SMPPConfig{
		SMSC:       viper.GetString("smpp_config.smsc"),
		SystemID:   viper.GetString("smpp_config.system_id"),
		Password:   viper.GetString("smpp_config.password"),
		SystemType: viper.GetString("smpp_config.system_type"),
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     viper.GetString("database_config.host"),
		Port:     viper.GetInt("database_config.port"),
		User:     viper.GetString("database_config.user"),
		Password: viper.GetString("database_config.password"),
		DBName:   viper.GetString("database_config.dbname"),
	}
}

func loadProvidersConfig() []Provider {
	providers := []Provider{}
	if err := viper.UnmarshalKey("providers", &providers); err != nil {
		log.Printf("Error unmarshaling providers config, %s", err)
	}
	
	for _, p := range providers {
		p.SMSC = p.Address + ":" + fmt.Sprintf("%d", p.Port)
	}

	return providers
}
