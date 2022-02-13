package repository

import (
	"context"

	"github.com/tragicpixel/fruitbar/pkg/models"
)

// TODO: Remove context

// User provides an interface for performing operations on a repository of user accounts.
type User interface {
	// Fetch find and returns an array of all users in the repository. Returns nil on error.
	Fetch(ctx context.Context, num int64) ([]*models.User, error)
	// GetByID finds and returns an individual user with the supplied id. Returns nil on error.
	GetByID(ctx context.Context, num int64) (*models.User, error)
	// GetByID finds and returns an individual user with the supplied username. Returns nil on error.
	GetByUsername(uname string) (*models.User, error)
	// Create creates a new order and places it in the repository. Returns the ID of the newly created order, -1 on error.
	Create(ctx context.Context, u *models.User) (int64, error)
	// Update updates an existing user in the repository. Returns nil on error.
	Update(ctx context.Context, u *models.User) (*models.User, error)
	// Delete removes an existing user with the supplied id from the repository. Returns true on success, false on error.
	Delete(ctx context.Context, id int64) (bool, error)
	// HashPassword hashes the supplied password and updates the supplied user's password to the hashed version.
	// This is normally used when creating a new user.
	HashPassword(u *models.User, pass string) error
	// CheckPassword checks if the supplied user's *hashed* password matches the supplied raw (plain text) password.
	// This is normally used when logging in with an existing user.
	CheckPassword(u *models.User, rawPass string) error
}
