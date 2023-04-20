package config

import "github.com/joho/godotenv"

var Env map[string]string

func init() {
	e := loadEnvVariables()
	Env = e
}

func loadEnvVariables() map[string]string {
	// Load environment variables from .env file
	var env map[string]string
	var err error
	env, err = godotenv.Read()
	if err != nil {
		panic("env load error. ")
	}
	return env
}
