// Command users starts the users API service.
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

	// Unless you set these on your local machine, the service won't be able to run...but you should be running it in docker.
	connection := pgdriver.PostgresConnectionConfig{
		Host:     os.Getenv("FRUITBAR_DATABASE_SERVICE_NAME"),
		Port:     "5423",
		Database: "fruitbar",
		Username: "postgres",
		Password: "fruitbar",
	}

	FruitbarUsersService, err := service.NewUsersService(8001, &connection)
	if err != nil {
		logrus.Error("failed to create the users service:" + err.Error())
		panic("failed to create the users service:" + err.Error())
	}
	logrus.Info("Server listening at port ", FruitbarUsersService.Port)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", FruitbarUsersService.Port), FruitbarUsersService.Router))
}
