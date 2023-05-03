package config

import (
	"github.com/joho/godotenv"
)

func LoadEnvVariables(filenames ...string) (map[string]string, error) {
	var env map[string]string
	var err error
	env, err = godotenv.Read(filenames...)
	if err != nil {
		return nil, err
	}
	return env, nil
}
