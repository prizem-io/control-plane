package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type (
	// Config represents the top level config.
	Config struct {
		Database DBConfig         `mapstructure:"database"`
		HTTP     HTTPServerConfig `mapstructure:"http"`
	}

	// DBConfig contains the fields needed to connect to the RDBMS.
	DBConfig struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Name     string `mapstructure:"name"`
		Username string `mapstructure:"user"`
		Password string `mapstructure:"password"`
	}

	// HTTPServerConfig contains the config for HTTP server
	HTTPServerConfig struct {
		Port int `mapstructure:"port"`
	}
)

const (
	envEnvironment    = "PRIZEM_ENV"
	envWorkingDir     = "PRIZEM_DIR"
	defaultConfigFile = "config"
)

var (
	env = os.Getenv(envEnvironment)
	err error
)

// LoadConfig sets up viper and returns the application configuration
func LoadConfig() (*Config, error) {
	var home string
	if wd := os.Getenv(envWorkingDir); wd != "" {
		home = wd
		if !strings.HasSuffix(home, "/") {
			home = home + "/"
		}
	}

	var configName = "config"
	if env != "" {
		configName += "." + env
	}

	viper.SetConfigName(configName)
	viper.AddConfigPath(home + "etc/")
	viper.SetEnvPrefix("PRIZEM")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err = viper.ReadInConfig()

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// ConnectDB connects to a Postgres database.
func ConnectDB(dbConfig *DBConfig) (*sqlx.DB, error) {
	connect := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.Name, dbConfig.Username, dbConfig.Password)
	return sqlx.Connect("postgres", connect)
}
