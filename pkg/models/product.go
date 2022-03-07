package models

import (
	"errors"
	"unicode/utf8"

	"gorm.io/gorm"
)

// swagger:model product
// Product holds information about a given product.
type Product struct {
	gorm.Model
	Name string `json:"name"`
	// The emoji used to represent the product. (if applicable)
	Symbol     string  `json:"symbol"`
	Price      float64 `json:"price"`
	NumInStock int     `json:"numInStock"`
}

func ValidateProduct(product *Product) (bool, error) {
	_, err := ValidateProductSymbol(product.Symbol)
	if err != nil {
		return false, errors.New("failed to validate product: " + err.Error())
	}
	_, err = ValidateProductPrice(product.Price)
	if err != nil {
		return false, errors.New("failed to validate product: " + err.Error())
	}
	return true, nil
}

// ValidatePartialProductUpdate validates the supplied selected fields of the supplied product.
func ValidatePartialProductUpdate(product *Product, selectedFields []string) (bool, error) {
	var err error
	// this is not very maintainable in the long run, your options are:
	// write a custom json.Marshal method
	// use code generation tools to extract the names in the json annotation
	for _, field := range selectedFields {
		switch field {
		case "symbol":
			_, err = ValidateProductSymbol(product.Symbol)
		case "price":
			_, err = ValidateProductPrice(product.Price)
		case "numInStock":
			_, err = ValidateProductNumInStock(product.NumInStock)
		}
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func ValidateProductSymbol(symbol string) (bool, error) {
	length := utf8.RuneCountInString(symbol)
	if length > 1 {
		return false, errors.New("symbol must be exactly 1 rune")
	} else if length < 1 {
		return false, errors.New("symbol must be exactly 1 rune")
	} else {
		return true, nil
	}
}

func ValidateProductPrice(price float64) (bool, error) {
	if price <= 0 {
		return false, errors.New("price must be greater than zero")
	} else {
		return true, nil
	}
}

func ValidateProductNumInStock(numInStock int) (bool, error) {
	if numInStock < 0 {
		return false, errors.New("numInStock must be greater than or equal to zero")
	} else {
		return true, nil
	}
}
