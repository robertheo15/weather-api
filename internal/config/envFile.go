package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvFile() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file %s", err.Error())
	} else {
		log.Println("Success load .env file")
	}
}
