package repository

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
)

// Order provides an interface for performing operations on a repoistory of orders.
type Order interface {
	// Fetch find and returns an array of all orders in the repository. Returns nil on error.
	Fetch(num int64) ([]*models.FruitOrder, error)
	// GetByID finds and returns an individual order with the supplied id. Returns nil on error.
	GetByID(num int64) (*models.FruitOrder, error)
	// Create creates a new order and places it in the repository. Returns the ID of the newly created order, -1 on error.
	Create(u *models.FruitOrder) (int64, error)
	// Update updates an existing order in the repository. Returns nil on error.
	Update(u *models.FruitOrder) (*models.FruitOrder, error)
	// Delete removes an order with the supplied id from the repository. Returns true on success, false on error.
	Delete(id int64) (bool, error)
}
