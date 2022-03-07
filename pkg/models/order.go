package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// CreditCardInfo holds the credit card information needed to run a credit card.
type CreditCardInfo struct {
	// Credit card number.
	Number string `json:"number"`
	// Name on the card.
	CardholderName string `json:"cardholdername"`
	// Expiration date. (mm/yy format)
	ExpirationDate string `json:"expirationdate"`
	// Zipcode on the card. (5 or 10 digits)
	Zipcode string `json:"zipcode"`
	// CVV on the card. (3 or 4 digits)
	Cvv string `json:"cvv"`
}

// PaymentInfo holds the payment information about a given order.
type PaymentInfo struct {
	// Whether the order paid in cash. (if false, this means it is paid with a credit card)
	Cash bool `json:"cash"`
	// Credit card info for this order.
	CardInfo CreditCardInfo `json:"cardinfo" gorm:"embedded"`
}

// swagger:model order
// Order holds all the information in an order of fruit.
type Order struct {
	gorm.Model
	// User ID of the user who owns the order.
	OwnerID uint `json:"ownerid"`
	// All of the items present in the order.
	Items []Item `json:"items"`
	// Payment information for this order.
	PaymentInfo PaymentInfo `json:"paymentinfo" gorm:"embedded"`
	// Tax rate for this order.
	TaxRate float64 `json:"taxrate"`
	// Subtotal of the order. (before tax+tip)
	Subtotal float64 `json:"subtotal"`
	// Tax on the order.
	Tax float64 `json:"tax"`
	// Total cost of the order.
	Total float64 `json:"total"`
}

// ValidateCreditCardExpirationDate determines whether a credit card's expiration date is valid. (4 digit mm/yy string)
func ValidateCreditCardExpirationDate(expDate *string) (bool, error) {
	return true, nil
}

// ValidateCreditCardNumber determines whether a supplied credit card number is valid. (16 digit string)
func ValidateCreditCardNumber(cardNumber *string) (bool, error) {
	return true, nil
}

// ValidateCreditCardCVV determines whether a supplied CVV is valid. (3-4 digit string)
func ValidateCreditCardCVV(cvv *string) (bool, error) {
	return true, nil
}

// ValidateZipcode determines whether a supplied zipcode is valid. (5 or 10 digit string)
func ValidateZipcode(zipcode *string) (bool, error) {
	return true, nil
}

// ValidateCreditCardInfo determines whether all of the supplied credit card information is valid.
func ValidateCreditCardInfo(cardInfo *CreditCardInfo) (bool, error) {
	expDateIsValid, expDateError := ValidateCreditCardExpirationDate(&cardInfo.ExpirationDate)
	numberIsValid, numberError := ValidateCreditCardNumber(&cardInfo.Number)
	cvvIsValid, cvvError := ValidateCreditCardCVV(&cardInfo.Cvv)
	zipcodeIsValid, zipcodeError := ValidateZipcode(&cardInfo.Zipcode)
	if expDateIsValid && numberIsValid && cvvIsValid && zipcodeIsValid {
		return true, nil
	} else {
		errorMsgs := []string{expDateError.Error(), numberError.Error(), cvvError.Error(), zipcodeError.Error()}
		msg := strings.Join(errorMsgs, ", ")
		return false, errors.New(msg)
	}
}

// ValidateFruitOrderPaymentInfo determines whether all of the supplied payment information for an order is valid.
func ValidateFruitOrderPaymentInfo(info *PaymentInfo) (bool, error) {
	if info.Cash {
		return true, nil
	} else {
		return ValidateCreditCardInfo(&info.CardInfo)
	}
}

// ValidateFruitOrderId validates whether the supplied order's ID is valid. (if it exists)
func ValidateFruitOrderId(order *Order) (bool, error) {
	if order.ID == 0 { // 0 considered empty by go for an int
		return false, errors.New("ID cannot be null")
	}
	return true, nil
}

// ValidateFruitOrder validates whether the supplied order is valid. (totals, payment info, and id need to be valid)
func ValidateFruitOrder(order *Order) (bool, error) {
	paymentInfoIsValid, paymentInfoError := ValidateFruitOrderPaymentInfo(&order.PaymentInfo)
	idIsValid, idError := ValidateFruitOrderId(order)
	totalIsValid := order.validateTotal()
	if paymentInfoIsValid && idIsValid && totalIsValid {
		return true, nil
	} else {
		errorMsgs := []string{paymentInfoError.Error(), idError.Error()}
		msg := strings.Join(errorMsgs, ", ")
		return false, errors.New(msg)
	}
}

// ValidateNewFruitOrder validates whether the supplied new fruit order (freshly created) is valid. (payment info needs to be valid)
func ValidateNewFruitOrder(order *Order) (bool, error) {
	// Need also to validate that subtotal, tax, and total are empty??
	subtotalIsValid, taxIsValid, totalIsValid := true, true, true
	var subtotalError, taxError, totalError error
	if order.Subtotal != 0.0 {
		subtotalIsValid = false
		subtotalError = errors.New("subtotal must be empty")
	}
	if order.Tax != 0.0 {
		taxIsValid = false
		taxError = errors.New("tax must be empty")
	}
	if order.Total != 0.0 {
		totalIsValid = false
		totalError = errors.New("total must be empty")
	}
	paymentInfoIsValid, paymentInfoError := ValidateFruitOrderPaymentInfo(&order.PaymentInfo)
	if subtotalIsValid && taxIsValid && totalIsValid && paymentInfoIsValid {
		return true, nil
	} else {
		errorMsgs := []string{subtotalError.Error(), taxError.Error(), totalError.Error(), paymentInfoError.Error()}
		msg := strings.Join(errorMsgs, ", ")
		return false, errors.New(msg)
	}
}

func (o *Order) validateTax() bool {
	return (o.Tax == o.Subtotal*o.TaxRate)
}

func (o *Order) validateTotal() bool {
	return (o.Total == o.Subtotal+o.Tax)
}

func ValidateOrderUpdate(order *Order, selectedFields []string) (bool, error) {
	var err error
	valid := true
	for _, field := range selectedFields {
		switch field {
		case "ownerid":
			_, err = ValidateFruitOrderId(order)
		case "paymentinfo":
			_, err = ValidateFruitOrderPaymentInfo(&order.PaymentInfo)
		case "items":
		case "taxrate":
		}
		if err != nil || !valid {
			return false, err
		}
	}
	return true, nil
}
