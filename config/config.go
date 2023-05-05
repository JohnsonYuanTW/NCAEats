package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

// LoadEnvVariables reads environment variables from the specified files and returns them as a map.
// It takes one or more filenames as input parameters and returns a map of key-value pairs representing the environment variables.
// If an error occurs while reading the files, it returns nil and the error message.
func LoadEnvVariables(filenames ...string) (map[string]string, error) {
	// Read the environment variables from the specified files using the godotenv package.
	env, err := godotenv.Read(filenames...)

	// If an error occurred while reading the files, return nil and the error message.
	if err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %v", err)
	}

	// Otherwise, return the map of environment variables.
	return env, nil
}
