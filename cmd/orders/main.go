// Command orders starts the orders API service.
package main

import (
	"fmt"
	"net/http"
	"os"

	pgdriver "github.com/tragicpixel/fruitbar/pkg/driver/postgres"
	"github.com/tragicpixel/fruitbar/pkg/service"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	connection := pgdriver.PostgresConnectionConfig{
		Host:     os.Getenv("FRUITBAR_DATABASE_SERVICE_NAME"),
		Port:     "5423",
		Database: "fruitbar",
		Username: "postgres",
		Password: "fruitbar",
	}

	// TODO: Wait time + Retry count for connecting to DB, don't just immediately fail.
	FruitBarOrdersService, err := service.NewOrdersService(8000, &connection)
	if err != nil {
		logrus.Error("failed to create the orders service:" + err.Error())
		panic("failed to create the orders service:" + err.Error())
	}
	logrus.Info("Server listening at port ", FruitBarOrdersService.Port)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", FruitBarOrdersService.Port), FruitBarOrdersService.Router))
}
