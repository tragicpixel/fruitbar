package utils

import (
	"errors"
	"os"
)

// GetEnv returns the value of the environment variable with the supplied name.
// If no variable with that name is set, returns an empty string and an error.
func GetEnv(name string) (string, error) {
	value, exists := os.LookupEnv(name)
	if exists {
		return value, nil
	} else {
		return "", errors.New("Environment variable '" + name + "' is not set")
	}
}
