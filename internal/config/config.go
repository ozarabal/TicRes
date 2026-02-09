package config

import "github.com/spf13/viper"

type Config struct {
	Server ServerConfig
	DB     DatabaseConfig
	JWT		JWTConfig
	Cache	RedisConfig
}

type ServerConfig struct {
	Port string
}

type JWTConfig struct{
	Secret 	string
	ExpTime int
}

type RedisConfig struct{
	Host  	string
	Port	string
	Password string
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
	viper.AutomaticEnv()

	// .env file is optional when environment variables are set directly
	_ = viper.ReadInConfig()

	var cfg Config
	
	// Mapping manual agar lebih aman
	cfg.Server.Port = viper.GetString("PORT")
	cfg.DB.Host = viper.GetString("DB_HOST")
	cfg.DB.Port = viper.GetString("DB_PORT")
	cfg.DB.User = viper.GetString("DB_USER")
	cfg.DB.Password = viper.GetString("DB_PASSWORD")
	cfg.DB.Name = viper.GetString("DB_NAME")
	cfg.JWT.Secret = viper.GetString("JWT_SECRET")
	cfg.JWT.ExpTime = viper.GetInt("JWT_EXP_TIME")
	cfg.Cache.Host = viper.GetString("CACHE_HOST")
	cfg.Cache.Password = viper.GetString("CACHE_PASSWORD")
	cfg.Cache.Port = viper.GetString("CACHE_PORT")

	return &cfg, nil
}