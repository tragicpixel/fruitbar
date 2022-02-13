package main

import (
	"github.com/tragicpixel/fruitbar/pkg/service"

	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{}) // use this wherever you are aggregating logs
	//logrus.SetFormatter(&logrus.TextFormatter{}) // use this when developing stuff that doesn't involve changing log output format etc
	FruitbarUsersService, err := service.NewUsersService(8001)
	if err != nil {
		logrus.Error("failed to create the users service:" + err.Error())
		panic("failed to create the users service:" + err.Error())
	}

	logrus.Info("Server listening at port ", FruitbarUsersService.Port)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", FruitbarUsersService.Port), FruitbarUsersService.Router))
}
