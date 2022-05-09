package models

import (
	"errors"

	"gorm.io/gorm"
)

// swagger:model item
type Item struct {
	gorm.Model
	OrderID   uint `json:"orderid"`
	ProductID uint `json:"productid"`
	Quantity  int  `json:"quantity"`
}

func (i *Item) ValidateOrderID() error {
	if i.OrderID <= 0 {
		return errors.New("orderid must be greater than zero")
	}
	return nil
}

func (i *Item) ValidateProductID() error {
	if i.ProductID <= 0 {
		return errors.New("productid must be greater than zero")
	}
	return nil
}

func (i *Item) ValidateQuantity() error {
	if i.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	return nil
}
