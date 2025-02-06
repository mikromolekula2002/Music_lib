package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBHost         string `mapstructure:"DB_HOST"`
	DBPort         string `mapstructure:"DB_PORT"`
	DBUser         string `mapstructure:"DB_USER"`
	DBPassword     string `mapstructure:"DB_PASSWORD"`
	DBName         string `mapstructure:"DB_NAME"`
	ServerPort     string `mapstructure:"SERVER_PORT"`
	LoggerLevel    string `mapstructure:"LOGGER_LEVEL"`
	LoggerOut      string `mapstructure:"LOGGER_OUT"`
	LoggerFilePath string `mapstructure:"LOGGER_FILEPATH"`
	EnvType        string `mapstructure:"ENV_TYPE"`
	MusicAPIHost   string `mapstructure:"MUSIC_API_HOST"`
	MusicBaseURL   string `mapstructure:"MUSIC_BASE_URL"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
