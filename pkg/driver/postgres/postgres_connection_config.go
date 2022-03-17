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

const (
	// Name of environment variable containing the database hostname.
	databaseHostnameEnv = "FRUITBAR_DB_HOSTNAME"
	// Name of environment variable containing the database connection port.
	databasePortEnv = "FRUITBAR_DB_PORT"
	// Name of environment variable containing the database database name.
	databaseDBNameEnv = "FRUITBAR_DB_DATABASE"
	// Name of environment variable containing the database user to use for transactions.
	databaseUsernameEnv = "FRUITBAR_DB_USER"
	// Name of environment variable containing the password for the database user.
	databasePasswordEnv = "FRUITBAR_DB_PASSWORD"
)

// NewPostgresConnectionConfigFromEnv returns a new connection configuration based on the values of environment variables.
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
