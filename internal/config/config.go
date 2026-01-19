package config

import "github.com/spf13/viper"

type Config struct {
	Server ServerConfig
	DB     DatabaseConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// LoadConfig membaca file .env dan memasukkannya ke struct Config
func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	
	// Mapping manual agar lebih aman
	cfg.Server.Port = viper.GetString("PORT")
	cfg.DB.Host = viper.GetString("DB_HOST")
	cfg.DB.Port = viper.GetString("DB_PORT")
	cfg.DB.User = viper.GetString("DB_USER")
	cfg.DB.Password = viper.GetString("DB_PASSWORD")
	cfg.DB.Name = viper.GetString("DB_NAME")

	return &cfg, nil
}