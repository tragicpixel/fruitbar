package models

import "gorm.io/gorm"

// swagger:model user
// User holds information about a given user account.
type User struct {
	gorm.Model
	Name     string `json:"name"`
	Password string `json:"password"`
}
