package repository

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/utils"
)

// Order provides an interface for performing operations on a repoistory of orders.
type Product interface {
	// Fetch find and returns an array of all products in the repository. Returns nil on error.
	Fetch(pageSeekOptions *utils.PageSeekOptions) ([]*models.Product, error)
	// Exists determines if a product with the supplied id exists.
	Exists(id uint) (bool, error)
	// GetByID finds and returns an individual product with the supplied id. Returns nil on error.
	GetByID(id uint) (*models.Product, error)
	// Create creates a new product and places it in the repository. Returns the ID of the newly created product, -1 on error.
	Create(u *models.Product) (uint, error)
	// Update updates an existing product in the repository. Returns nil on error.
	Update(u *models.Product, fields []string) (*models.Product, error)
	// Delete removes a product with the supplied id from the repository. Returns true on success, false on error.
	Delete(id uint) (bool, error)
}
