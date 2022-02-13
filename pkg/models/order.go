package models

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

const (
	APPLE_PRICE_DOLLARS  = 0.5
	BANANA_PRICE_DOLLARS = 0.25
	ORANGE_PRICE_DOLLARS = 1.25
	CHERRY_PRICE_DOLLARS = 0.75
	SALES_TAX_PERCENT    = 0.05
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
// FruitOrder holds all the information in an order of fruit.
type FruitOrder struct {
	gorm.Model
	NumApples   int `json:"numApples"`
	NumBananas  int `json:"numBananas"`
	NumOranges  int `json:"numOranges"`
	NumCherries int `json:"numCherries"`
	// Payment information for this order.
	PaymentInfo PaymentInfo `json:"paymentinfo" gorm:"embedded"`
	// Subtotal of the order. (before tax+tip)
	Subtotal float64 `json:"subtotal"`
	// Tax on the order.
	Tax float64 `json:"tax"`
	// Tip left on the order.
	TipAmt float64 `json:"tipamt"`
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

// ValidateFruitOrderTipAmount determines whether the supplied tip amount is valid.
// Change to accept int?
func ValidateFruitOrderTipAmount(order *FruitOrder) bool {
	return order.TipAmt >= 0.0
}

// ValidateFruitOrderTotals determines whether the supplied order's subtotal, tax, tip, and total are valid.
func ValidateFruitOrderTotals(order *FruitOrder) (bool, error) {
	expectedSubtotal := CalculateFruitOrderSubtotal(order)
	if expectedSubtotal != order.Subtotal {
		return false, errors.New("invalid subtotal")
	}
	expectedTax := CalculateFruitOrderTax(order)
	if expectedTax != order.Tax {
		return false, errors.New("invalid tax")
	}
	if !ValidateFruitOrderTipAmount(order) {
		return false, errors.New("invalid tip amount")
	}
	expectedTotal := CalculateFruitOrderTotal(order)
	if order.Total != expectedTotal {
		return false, errors.New("invalid total")
	}
	return true, nil
}

// ValidateFruitOrderId validates whether the supplied order's ID is valid. (if it exists)
func ValidateFruitOrderId(order *FruitOrder) (bool, error) {
	if order.ID == 0 { // 0 considered empty by go for an int
		return false, errors.New("ID cannot be null")
	}
	return true, nil
}

// ValidateFruitOrder validates whether the supplied order is valid. (totals, payment info, and id need to be valid)
func ValidateFruitOrder(order *FruitOrder) (bool, error) {
	totalsAreValid, totalsError := ValidateFruitOrderTotals(order)
	paymentInfoIsValid, paymentInfoError := ValidateFruitOrderPaymentInfo(&order.PaymentInfo)
	idIsValid, idError := ValidateFruitOrderId(order)
	if totalsAreValid && paymentInfoIsValid && idIsValid {
		return true, nil
	} else {
		errorMsgs := []string{totalsError.Error(), paymentInfoError.Error(), idError.Error()}
		msg := strings.Join(errorMsgs, ", ")
		return false, errors.New(msg)
	}
}

// ValidateNewFruitOrder validates whether the supplied new fruit order (freshly created) is valid. (payment info needs to be valid)
func ValidateNewFruitOrder(order *FruitOrder) (bool, error) {
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

// CalculateFruitOrderSubtotal returns the correct subtotal based on the supplied order.
func CalculateFruitOrderSubtotal(order *FruitOrder) float64 {
	subtotal := 0.0
	subtotal = subtotal + float64(order.NumApples)*APPLE_PRICE_DOLLARS
	subtotal = subtotal + float64(order.NumBananas)*BANANA_PRICE_DOLLARS
	subtotal = subtotal + float64(order.NumOranges)*ORANGE_PRICE_DOLLARS
	subtotal = subtotal + float64(order.NumCherries)*CHERRY_PRICE_DOLLARS
	return subtotal
}

// CalculateFruitOrderTax returns the correct tax amount based on the supplied order.
func CalculateFruitOrderTax(order *FruitOrder) float64 {
	tax := CalculateFruitOrderSubtotal(order) * SALES_TAX_PERCENT
	return tax
}

// CalculateFruitOrderTotal returns the correct total amount based on the supplied order.
// It will recalculate the subtotal and tax; if those values are incorrect, the total may not match them.
func CalculateFruitOrderTotal(order *FruitOrder) float64 {
	total := CalculateFruitOrderSubtotal(order) + CalculateFruitOrderTax(order) + order.TipAmt
	return total
}
