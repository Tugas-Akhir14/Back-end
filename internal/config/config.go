package config

import "github.com/spf13/viper"

type Config struct {
    DatabaseURL string
    JWTSecret   string
}

func LoadConfig() Config {
    viper.SetConfigFile(".env")
    err := viper.ReadInConfig()
    if err != nil {
        panic("Error reading config file: " + err.Error())
    }

    return Config{
        DatabaseURL: viper.GetString("DATABASE_URL"),
        JWTSecret:   viper.GetString("JWT_SECRET"),
    }
}
