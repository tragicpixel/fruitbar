package repository

import (
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/utils"
)

// Order provides an interface for performing operations on a repoistory of orders.
type Item interface {
	// Fetch find and returns an array of all items in the repository. Returns nil on error.
	Fetch(pageSeekOptions utils.PageSeekOptions) ([]*models.Item, error)
	// Exists determines if an item with the supplied id exists.
	Exists(id int) (bool, error)
	// GetByID finds and returns an individual item with the supplied id. Returns nil on error.
	GetByID(id uint) (*models.Item, error)
	// GetByOrderID returns an array of all the items in the repository with the supplied order id. Returns nil on error.
	GetByOrderID(id uint) ([]*models.Item, error)
	// Create creates a new item and places it in the repository. Returns the ID of the newly created item, -1 on error.
	Create(i *models.Item) (uint, error)
	// Update updates an existing item in the repository. Returns nil on error.
	Update(i *models.Item, fields []string) (*models.Item, error)
	// Delete removes an item with the supplied id from the repository. Returns true on success, false on error.
	Delete(id int64) (bool, error)
}
