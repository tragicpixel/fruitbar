package models

import jwt "github.com/dgrijalva/jwt-go"

// JwtWrapper holds vital information about a given JSON web token for the fruitbar application.
type JwtWrapper struct {
	SecretKey       string
	Issuer          string
	ExpirationHours int64
}

// JwtClaim holds the standard JWT claim in addition to any other claims made that will need to be verified by the fruitbar application.
type JwtClaim struct {
	jwt.StandardClaims
	Username string // TODO: just put the user id and lookup the username/role as needed from the database instead of putting it in the claims? (more secure)
	Role     string
}
