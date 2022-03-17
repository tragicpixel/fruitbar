package postgres

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/tragicpixel/fruitbar/pkg/driver"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

// OpenConnection attempts to open a connection to a postgres database using gorm and returns a driver with a valid postgres connection.
func OpenConnection(connectionConfig *PostgresConnectionConfig) (*driver.DB, error) {
	zaplogger := zapgorm2.New(zap.L())
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		connectionConfig.Host, connectionConfig.Username, connectionConfig.Password, connectionConfig.Database,
	)
	logrus.Info("Opening connection to database: " + dbinfo) // may not want to include user credentials in the log file
	conn, err := gorm.Open(postgres.Open(dbinfo), &gorm.Config{Logger: zaplogger})
	if err != nil {
		logrus.Error("Error opening database connection: " + err.Error())
		return nil, err
	}
	logrus.Info("Successfully opened connection to the database.")
	db := driver.DB{Postgres: conn}
	return &db, nil
}

// SetupTables checks that the postgres database for the given database driver contains a table matching the schema gorm would create for that object.
// Optionally, it can initialize the matching table if it is missing from the database or structured incorrectly based on the current model.
func SetupTables(db *driver.DB, object interface{}, init bool) error {
	stmt := &gorm.Statement{DB: db.Postgres}
	err := stmt.Parse(object)
	if err != nil {
		msg := fmt.Sprintf("Failed to parse object %+v: %s", object, err.Error())
		logrus.Error(msg)
		return errors.New(msg)
	}
	tableName := stmt.Schema.Table
	if !db.Postgres.Migrator().HasTable(object) { // This will also pick up on if the schema in the DB does not match gorm's data model (based on the code)
		logrus.Info("Table not found: " + tableName)
		if init {
			// If the table exists, but not for the parsed object, this means the schema has changed, so drop the table.
			// In a real production application, you probably wouldn't want to do this and instead just fail outright, or perform some kind of data migration instead.
			err := db.Postgres.Migrator().DropTable(object)
			if err != nil {
				msg := fmt.Sprintf("Failed to drop table %s: %s", tableName, err.Error())
				logrus.Error(msg)
				return errors.New(msg)
			}
			err = db.Postgres.Migrator().CreateTable(object)
			if err != nil {
				msg := fmt.Sprintf("Failed to create table %s: %s", tableName, err.Error())
				logrus.Error(msg)
				return errors.New(msg)
			}
			logrus.Info("Created table: " + tableName)
		} else {
			msg := "Couldn't find table " + tableName
			logrus.Error(msg)
			return errors.New(msg)
		}
	}
	return nil
}
