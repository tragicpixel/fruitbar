package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"gorm.io/gorm"
)

// swagger:model user
// User holds information about a given user account.
type User struct {
	gorm.Model
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// ValidateUser checks if a User object is valid.
func ValidateUser(user *User) (bool, error) {
	_, err := ValidateUserName(user.Name)
	if err != nil {
		return false, err
	}
	_, err = ValidatePassword(user.Password)
	if err != nil {
		return false, err
	}
	_, err = ValidateRole(user.Role)
	if err != nil {
		return false, err
	}
	return true, nil
}

func getUsernameMinLength() int { return 1 }
func getUsernameMaxLength() int { return 100 }

// ValidateUserName checks if the supplied username is valid.
// TODO: this should probably just be private, and the larger validate function be public
func ValidateUserName(name string) (bool, error) {
	length := len(name)
	if length > getUsernameMaxLength() {
		return false, errors.New("name must be less than " + strconv.Itoa(getUsernameMaxLength()) + " characters long")
	} else if length < getUsernameMinLength() {
		return false, errors.New("name must be at least " + strconv.Itoa(getUsernameMinLength()) + " characters long")
	} else {
		return true, nil
	}
}

func getPasswordMinLength() int         { return 8 }
func getPasswordMaxLength() int         { return 100 }
func getValidSpecialCharacters() string { return "!@#$%^&*()-_=[];',./?~`" }

// GetPasswordFormatMessage returns a string containing information about the constraints applied when validating a password.
// The goal of having this function is to keep information about the password constraints limited to a single source of truth: directly in the code, so it's always up to date.
func GetPasswordFormatMessage() string {
	msg := "Password must be between " + strconv.Itoa(getPasswordMinLength()) + " and " + strconv.Itoa(getPasswordMaxLength()) + " characters long, "
	msg += "contain at least one digit, "
	msg += "contain at least one of the following special characters: " + getValidSpecialCharacters()
	return msg
}

func ValidatePassword(pass string) (bool, error) {
	length := len(pass)
	maxLength, minLength := getPasswordMaxLength(), getPasswordMinLength()
	if length > maxLength {
		return false, errors.New("password must be less than " + strconv.Itoa(maxLength) + " characters long")
	} else if length < minLength {
		return false, errors.New("password must be at least " + strconv.Itoa(minLength) + " characters long")
	}

	containsDigit := false
	for _, char := range pass {
		if unicode.IsDigit(char) {
			containsDigit = true
			break
		}
	}
	if !containsDigit {
		return false, errors.New("password must contain at least one digit")
	}

	specialCharacters := getValidSpecialCharacters()
	if !strings.ContainsAny(pass, specialCharacters) {
		return false, errors.New("password must contain at least one of the following special characters: " + specialCharacters)
	}

	return true, nil
}

// HasRole determines if a user's current role matches the access level of the supplied role.
// TODO: Rename this to HasAccessLevel ??
// TODO: make this a function for the *User object only
func HasRole(userRole string, role string) (bool, error) {
	switch role {
	case GetAdminRoleId():
		if userRole == GetAdminRoleId() {
			return true, nil
		}
		return false, nil
	case GetEmployeeRoleId():
		switch userRole {
		case GetAdminRoleId():
			return true, nil
		case GetEmployeeRoleId():
			return true, nil
		default:
			return false, nil
		}
	case GetCustomerRoleId():
		return true, nil
	default:
		msg := fmt.Sprintf("the role you are comparing against is invalid, role must be one of the following: %s", strings.Join(getValidRoles(), ", "))
		return false, errors.New(msg)
	}
}

func GetCustomerRoleId() string { return "customer" }
func GetEmployeeRoleId() string { return "employee" }
func GetAdminRoleId() string    { return "admin" }
func getValidRoles() []string {
	return []string{GetCustomerRoleId(), GetEmployeeRoleId(), GetAdminRoleId()}
}

func isRoleValid(role string) bool {
	for _, validRole := range getValidRoles() {
		if role == validRole {
			return true
		}
	}
	return false
}

func ValidateRole(role string) (bool, error) {
	isValid := isRoleValid(role)
	if !isValid {
		return false, errors.New("role must be one of the following: " + strings.Join(getValidRoles(), ", "))
	} else {
		return true, nil
	}
}

// ValidatePartialProductUpdate validates the supplied selected fields of the supplied product.
func ValidatePartialUserUpdate(user *User, selectedFields []string) (bool, error) {
	var err error
	// this is not very maintainable in the long run, your options are:
	// write a custom json.Marshal method
	// use code generation tools to extract the names in the json annotation
	for _, field := range selectedFields {
		switch field {
		case "name":
			_, err = ValidateUserName(user.Name)
		case "password":
			_, err = ValidatePassword(user.Password)
		case "role":
			_, err = ValidateRole(user.Role)
		}
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
