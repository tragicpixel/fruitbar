// Package roles
package roles

import (
	"errors"
	"fmt"
	"strings"
)

const (
	Customer = "customer"
	Employee = "employee"
	Admin    = "admin"
)

// HasRole validates if the supplied user role has the privilege level of the supplied role.
func HasRole(userRole string, role string) (bool, error) {
	switch role {
	case Admin:
		if userRole == Admin {
			return true, nil
		}
		return false, nil
	case Employee:
		switch userRole {
		case Admin:
			return true, nil
		case Employee:
			return true, nil
		default:
			return false, nil
		}
	case Customer:
		return true, nil
	default:
		msg := fmt.Sprintf("role is invalid, expected one of: %s got %s", ValidRolesMsg(), role)
		return false, errors.New(msg)
	}
}

// ValidRoles returns a slice of valid role IDs.
func ValidRoles() []string {
	return []string{Customer, Employee, Admin}
}

func ValidRolesMsg() string {
	return strings.Join(ValidRoles(), ", ")
}

// IsValid determines if the supplied role is a valid role ID.
func IsValid(role string) (bool, error) {
	for _, validRole := range ValidRoles() {
		if role == validRole {
			return true, nil
		}
	}
	return false, fmt.Errorf("role is invalid, expected one of: %s got %s", ValidRolesMsg(), role)
}
