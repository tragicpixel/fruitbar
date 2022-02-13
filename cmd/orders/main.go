package main

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tragicpixel/fruitbar/pkg/service"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	FruitBarOrdersService, err := service.NewOrdersService(8000)
	if err != nil {
		logrus.Error("failed to create the orders service:" + err.Error())
		panic("failed to create the orders service:" + err.Error())
	}

	logrus.Info("Server listening at port ", FruitBarOrdersService.Port)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", FruitBarOrdersService.Port), FruitBarOrdersService.Router))
}
