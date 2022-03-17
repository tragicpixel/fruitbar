// Package driver provides an interface to connect different types of persistent storage, such as a database, file, or cache.
package driver

import "gorm.io/gorm"

// DB holds the connections to different possible types of databases.
type DB struct {
	Postgres *gorm.DB
	// Mongo *mgo.database
	// etc
}
