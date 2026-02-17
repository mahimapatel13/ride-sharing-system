package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseConfig DatabaseConfig
	RedisConfig    RedisConfig
	RabbitMQConfig RabbitMQConfig
	MaxWorkerCount int
}

type DatabaseConfig struct {
	Port         string
	User         string
	Password     string
	DatabaseName string
	Host         string
	Address      string
}

type RabbitMQConfig struct{
	URL string
}

type RedisConfig struct {
	Protocol int
	Password string
	DB       int
	Address  string
}

// LoadEnv loads the environment and file and configures the app using the env file
func LoadEnv() Config {
	log.Println("Reading .env file")
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error in loading .env file: %v", err)
	}
	dbConfig := loadDBConfig()
	redisConfig := loadRedisConfig()
	mqConfig := loadRabbitMQConfig()

	config := Config{
		DatabaseConfig: dbConfig,
		RedisConfig:    redisConfig,
		RabbitMQConfig: mqConfig,
		MaxWorkerCount: getInt(getEnvValue("MAX_WORKER_COUNT", "1"), 1),
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
		Address: addr,
		Port:    port,
	}

	return dbConfig
}

// Loads RabbitMQ Config
func loadRabbitMQConfig() RabbitMQConfig {
	url := getEnvValue("RABBITMQ_URL", "DEFAULT_RABBITMQ_URL")
	

	mqConfig:= RabbitMQConfig{
		URL: url,
	}

	return mqConfig
}
// Loads Redis Config
func loadRedisConfig() RedisConfig {
	pass := getEnvValue("REDIS_PASS", "DEFAULT_DB_PASS")
	db := getEnvValue("REDIS_DB", "DEFUALT_DB_NAME")
	addr := getEnvValue("REDIS_ADDR", "DEFAULT_DB_ADDR")
	prot := getEnvValue("REDIS_PROTOCOL", "DEFAULT_DB_PROTOCOL")

	redisConfig := RedisConfig{
		Password: pass,
		DB:       getInt(db, 0),
		Address:  addr,
		Protocol:    getInt(prot, 2),
	}

	return redisConfig
}

func getEnvValue(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return defaultValue
}

func getInt(val string, def int) int {
	p, err := strconv.Atoi(val)

	if err != nil {
		p = def
	}

	return p
}