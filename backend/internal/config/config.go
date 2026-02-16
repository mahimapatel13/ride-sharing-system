package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseConfig DatabaseConfig
	RedisConfig    RedisConfig
}

type DatabaseConfig struct {
	Port         string
	User         string
	Password     string
	DatabaseName string
	Host         string
	Address      string
}

type RedisConfig struct {
	Port         string
	User         string
	Password     string
	DatabaseName string
	Host         string
	Address      string
}

// LoadEnv loads the environment and file and configures the app using the env file
func LoadEnv() Config {
	log.Println("Reading .env file")
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error in loading .env file: %v", err)
	}
	dbConfig := loadDBConfig()

	config := Config{
		DatabaseConfig: dbConfig,
	}

	log.Println(config)
	return config
}

// Loads DB Config
func loadDBConfig() DatabaseConfig {
	user := getEnvValue("DB_USER", "DEFAULT_DB_USER")
	pass := getEnvValue("DB_PASS", "DEFAULT_DB_PASS")
	name := getEnvValue("DB_NAME", "DEFUALT_DB_NAME")
	host := getEnvValue("DB_HOST", "DEFAULT_DB_HOST")
	addr := getEnvValue("DB_ADDR", "DEFAULT_DB_ADDR")
	port := getEnvValue("DB_PORT", "DEFAULT_DB_PORT")

	dbConfig := DatabaseConfig{
		User:         user,
		Password:     pass,
		DatabaseName: name,
		Host:         host,


        Address:      addr,
		Port:         port,
	}

	return dbConfig
}

// Loads Redis Config
func loadRedisConfig() RedisConfig {
	user := getEnvValue("REDIS_USER", "DEFAULT_DB_USER")
	pass := getEnvValue("REDIS_PASS", "DEFAULT_DB_PASS")
	name := getEnvValue("REDIS_NAME", "DEFUALT_DB_NAME")
	host := getEnvValue("REDIS_HOST", "DEFAULT_DB_HOST")
	addr := getEnvValue("REDIS_ADDR", "DEFAULT_DB_ADDR")
	port := getEnvValue("REDIS_PORT", "DEFAULT_DB_PORT")

	redisConfig := RedisConfig{
		User:         user,
		Password:     pass,
		DatabaseName: name,
		Host:         host,
		Address:      addr,
		Port:         port,
	}

	return redisConfig
}


func getEnvValue(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return defaultValue
}