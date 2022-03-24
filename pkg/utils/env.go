package utils

import (
	"fmt"
	"os"
)

// GetEnv returns the value of the environment variable with the supplied name.
func GetEnv(name string) (string, error) {
	value, exists := os.LookupEnv(name)
	if exists {
		return value, nil
	} else {
		return "", fmt.Errorf("environment variable '%s' is not set", name)
	}
}
