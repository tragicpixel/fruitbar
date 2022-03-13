package repository

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/utils"
)

// Product provides an interface for performing operations on a repository of products.
type Product interface {
	// Count returns the count of all the records matching the supplied seek options.
	Count(seek *utils.PageSeekOptions) (count int64, err error)
	// Fetch returns the products in the repository matching the supplied seek options.
	Fetch(pageSeekOptions *utils.PageSeekOptions) ([]*models.Product, error)
	// Exists determines if a product with the supplied id exists.
	Exists(id uint) (bool, error)
	// GetByID returns the product with the supplied id, if it exists.
	GetByID(id uint) (*models.Product, error)
	// Create creates a new product and returns the ID of the newly created product.
	Create(p *models.Product) (uint, error)
	// Update updates an existing product in the repository and returns the updated product.
	Update(p *models.Product, fields []string) (*models.Product, error)
	// Delete removes a product with the supplied id from the repository.
	Delete(id uint) error
}
