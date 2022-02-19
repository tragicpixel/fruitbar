package postgres

// PostgresConnectionConfig holds the properties necessary to configure a connection to a postgres database.
type PostgresConnectionConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}
