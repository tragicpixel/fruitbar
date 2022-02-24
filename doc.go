// Package fruitbar Fruitbar API
//
// Allows access to an API for managing fruit orders, product listings, and users.
//
//   version: 1.0.0
//   title: Fruitbar Authentication Service
//  Schemes:
//    -http
//    -https
//  Host: localhost:8001
//  BasePath: /
//	Consumes:
//	  - application/json
//  Produces:
//    - application/json
// Security:
// - bearer
//
// SecurityDefinitions:
// bearer:
//    type: apiKey
//    in: header
//    name: authorization
//    bearerFormat: JWT
//    description: JSON Web Token
// swagger:meta
package main
