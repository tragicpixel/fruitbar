package repository

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/utils"
)

// User provides an interface for performing operations on a repository of user accounts.
type User interface {
	// Count returns the count of all the records matching the supplied seek options.
	Count(seek *utils.PageSeekOptions) (count int64, err error)
	// Fetch returns the users in the repository matching the supplied seek options.
	Fetch(pageSeekOptions *utils.PageSeekOptions) ([]*models.User, error)
	// Exists determines if a user with the supplied id exists.
	Exists(id uint) (bool, error)
	// GetByID finds and returns an individual user with the supplied id. Returns nil on error.
	GetByID(id uint) (*models.User, error)
	// GetByID finds and returns an individual user with the supplied username. Returns nil on error.
	GetByUsername(uname string) (*models.User, error)
	// Create creates a new user and places it in the repository. Returns the ID of the newly created user, -1 on error.
	Create(u *models.User) (uint, error)
	// Update updates an existing user in the repository. Returns nil on error.
	Update(u *models.User, fields []string) (*models.User, error)
	// Delete removes an existing user with the supplied id from the repository. Returns true on success, false on error.
	Delete(id uint) error
	// HashPassword hashes the supplied password and updates the supplied user's password to the hashed version.
	HashPassword(u *models.User, pass string) error
	// CheckPassword checks if the supplied user's *hashed* password matches the supplied raw (plain text) password.
	CheckPassword(u *models.User, rawPass string) error
}
