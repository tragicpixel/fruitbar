// Command orders starts the orders API service.
package main

import (
	"fmt"
	"net/http"

	pgdriver "github.com/tragicpixel/fruitbar/pkg/driver/postgres"
	"github.com/tragicpixel/fruitbar/pkg/service"

	"github.com/sirupsen/logrus"
)

const (
	servicePortEnv = ""

	databaseHostnameEnv = "FRUITBAR_DATABASE_SERVICE_NAME"
	databasePortEnv     = "PORT"
	databaseDBNameEnv   = ""
	databaseUsernameEnv = ""
	databasePasswordEnv = ""
)

func main() {
	// Configure logging
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Configure the service
	config := service.ProductsServiceConfig{
		Port: 8002,
	}
	connection := pgdriver.PostgresConnectionConfig{
		Host:     "localhost", // just for testing the API functionality, once that's ironed out, go back to docker method
		Port:     "5423",
		Database: "fruitbar",
		Username: "postgres",
		Password: "fruitbar",
	}
	config.DatabaseConnection = &connection

	// Start the service
	FruitBarProductsService, err := service.NewProductsService(&config)
	if err != nil {
		msg := "failed to create the products service:"
		logrus.Error(msg + err.Error())
		panic(msg + err.Error())
	}
	logrus.Info("Server listening at port ", FruitBarProductsService.Port)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", FruitBarProductsService.Port), FruitBarProductsService.Router))
}
