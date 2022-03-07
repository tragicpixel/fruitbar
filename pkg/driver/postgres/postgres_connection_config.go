package postgres

import (
	"errors"

	"github.com/tragicpixel/fruitbar/pkg/utils"
)

// PostgresConnectionConfig holds the properties necessary to configure a connection to a postgres database.
type PostgresConnectionConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

// Names of environment variables used to configure the database connection
const (
	databaseHostnameEnv = "FRUITBAR_DB_HOSTNAME"
	databasePortEnv     = "FRUITBAR_DB_PORT"
	databaseDBNameEnv   = "FRUITBAR_DB_DATABASE"
	databaseUsernameEnv = "FRUITBAR_DB_USER"
	databasePasswordEnv = "FRUITBAR_DB_PASSWORD"
)

// NewPostgresConnectionConfigFromEnv creates a new connection configuration based on environment variables.
func NewPostgresConnectionConfigFromEnv() (*PostgresConnectionConfig, error) {
	host, err := utils.GetEnv(databaseHostnameEnv)
	if err != nil {
		return nil, errors.New("Failed to set database hostname: " + err.Error())
	}
	port, err := utils.GetEnv(databasePortEnv)
	if err != nil {
		return nil, errors.New("Failed to set database port: " + err.Error())
	}
	database, err := utils.GetEnv(databaseDBNameEnv)
	if err != nil {
		return nil, errors.New("Failed to set database database name: " + err.Error())
	}
	username, err := utils.GetEnv(databaseUsernameEnv)
	if err != nil {
		return nil, errors.New("Failed to set database username: " + err.Error())
	}
	password, err := utils.GetEnv(databasePasswordEnv)
	if err != nil {
		return nil, errors.New("Failed to set database password: " + err.Error())
	}
	return &PostgresConnectionConfig{
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		Password: password,
	}, nil
}
