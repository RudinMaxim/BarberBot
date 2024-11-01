package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var Texts map[string]string

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, proceeding without it.")
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found. Creating a default config file.")
			createDefaultConfig()
		} else {
			log.Fatalf("Error reading config file: %v", err)
		}
	} else {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}

	viper.SetConfigName("texts")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	viper.AddConfigPath("./texts")
	viper.AddConfigPath("../texts")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading texts file: %v", err)
	}

	if err := viper.Unmarshal(&Texts); err != nil {
		log.Fatalf("Unable to decode texts into map: %v", err)
	}
}

func createDefaultConfig() {
	viper.SetDefault("node.mode", os.Getenv("NODE_MODE"))

	viper.SetDefault("telegram.token", os.Getenv("TELEGRAM_TOKEN"))

	viper.SetDefault("database.host", os.Getenv("DB_HOST"))
	viper.SetDefault("database.port", os.Getenv("DB_PORT"))
	viper.SetDefault("database.user", os.Getenv("DB_USER"))
	viper.SetDefault("database.password", os.Getenv("DB_PASSWORD"))
	viper.SetDefault("database.name", os.Getenv("DB_NAME"))
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", 3600)
	viper.SetDefault("cache.host", os.Getenv("REDIS_HOST"))

	if err := viper.SafeWriteConfig(); err != nil {
		log.Fatalf("Error writing default config file: %v", err)
	}

	log.Println("Default config file created. Please edit it with your settings and restart the application.")
	os.Exit(0)
}
