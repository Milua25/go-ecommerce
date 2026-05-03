package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr      string
	Port      string
	ENV       string
	DBConfig  DBConfig
	JWTConfig JWTConfig
}
type DBConfig struct {
	Host              string
	Port              string
	User              string
	Password          string
	DBName            string
	UserCollection    string
	ProductCollection string
	// MaxOpenConns int
	// MaxIdleConns int
	// MaxIdleTime  string
}

type JWTConfig struct {
	SecretKey     string
	ExpiresIn     string
	RefreshSecret string
	RefreshExp    string
}

func loadEnv() {

	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v\n", err)
	}

}

func LoadConfig() (*Config, []string) {
	loadEnv()

	missingVars := make([]string, 0)

	requiredVars := func(key string) string {
		value := os.Getenv(key)
		if value == "" {
			log.Printf("Warning: Environment variable %s is not set\n", key)
			missingVars = append(missingVars, key)
		}
		return value
	}

	optionalVars := func(key, fallback string) string {
		value := os.Getenv(key)
		if value == "" {
			log.Printf("Info: Optional environment variable %s is not set, using fallback value\n", key)
			return fallback
		}
		return value
	}

	cfg := Config{
		Addr: optionalVars("ADDR", "0.0.0.0"),
		Port: optionalVars("PORT", "8000"),
		ENV:  optionalVars("ENV", "development"),
		DBConfig: DBConfig{
			Host:              requiredVars("DB_HOST"),
			Port:              requiredVars("DB_PORT"),
			User:              requiredVars("DB_USER"),
			Password:          requiredVars("DB_PASSWORD"),
			DBName:            requiredVars("DB_NAME"),
			UserCollection:    requiredVars("DB_USER_COLLECTION"),
			ProductCollection: requiredVars("DB_PRODUCT_COLLECTION"),
		},
		JWTConfig: JWTConfig{
			SecretKey:     requiredVars("JWT_SECRET_KEY"),
			ExpiresIn:     requiredVars("JWT_EXPIRES_IN"),
			RefreshSecret: requiredVars("JWT_REFRESH_SECRET"),
			RefreshExp:    requiredVars("JWT_REFRESH_EXPIRES_IN"),
		},
	}

	return &cfg, missingVars
}
