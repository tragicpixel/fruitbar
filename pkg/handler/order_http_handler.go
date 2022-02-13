package handler

import (
	"github.com/tragicpixel/fruitbar/pkg/driver"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	orderrepo "github.com/tragicpixel/fruitbar/pkg/repository/order"
	"github.com/tragicpixel/fruitbar/pkg/utils"

	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Order represents an http handler for performing operations on a repository of orders.
type Order struct {
	repo repository.Order
}

// NewOrderHandler creates a new http handler for performing operations on a repository of orders.
func NewOrderHandler(db *driver.DB) *Order {
	return &Order{
		repo: orderrepo.NewPostgresOrderRepo(db.Postgres), // this is where it is decided which implementation(/database type) of the Order Repo we will use
	}
}

// CreateOrder creates a new order in the repo based on the supplied HTTP request and sends a response in JSON to the user based on success or failure.
// Requires at least 1 fruit to be purchased, paymentInfo.cash must be true or paymentInfo.cardInfo must be filled out and valid.
// Any supplied values for subtotal, tax, and total, will be overwritten.
func (h *Order) CreateOrder(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodPost, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf("Create API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}

	var response = utils.JsonResponse{}
	var order models.FruitOrder
	err := utils.DecodeJSONBody(w, r, &order, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if err != nil {
		var request *utils.MalformedRequestError
		if errors.As(err, &request) {
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: request.Status, Message: request.Message}}
		} else {
			msg := "Failed to decode JSON body: " + err.Error()
			logrus.Error(msg)
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: msg}}
		}
	}

	if response.Error == nil {
		newOrderIsValid, newOrderValidationError := models.ValidateNewFruitOrder(&order)
		if !newOrderIsValid {
			msg := "Failed to validate new fruit order: " + newOrderValidationError.Error()
			logrus.Error(msg)
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: msg}}
		} else {
			// Calculate subtotal, tax, and total -> This will overwrite any supplied values
			order.Subtotal = models.CalculateFruitOrderSubtotal(&order)
			order.Tax = models.CalculateFruitOrderTax(&order)
			order.Total = models.CalculateFruitOrderTotal(&order)
			logrus.Info("Calculated order subtotal, tax, total")
			logrus.Info(fmt.Sprintf("Trying to insert new FruitOrder: %+v", order))
			createdId, err := h.repo.Create(&order)
			if err != nil {
				logrus.Error(fmt.Sprintf("Error inserting Fruit Order %+v into database: %s", order, err.Error()))
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create new order."}}
			} else {
				logrus.Info(fmt.Sprintf("Successfully inserted new fruit order: %+v", order))
				response = utils.JsonResponse{Data: []*models.FruitOrder{&order}, Id: strconv.Itoa(int(createdId))}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else {
		utils.WriteJSONResponse(w, http.StatusCreated, response)
	}
}

// GetOrders retrieves an existing order in the repo based on the supplied ID query parameter and returns a response in JSON containing either the order encoded in JSON or an error message.
func (h *Order) GetOrders(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodGet, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf("Read API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}

	var response = utils.JsonResponse{}
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		//response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "'id' parameter must be specified"}}
		// TODO: include options for pagination of the entire list of orders
		// count of total orders???
		logrus.Info("Retrieving all orders (max 100)...")
		var orders []*models.FruitOrder
		orders, err := h.repo.Fetch(100)
		if err != nil {
			logrus.Error(fmt.Sprintf("Error retrieving all orders: %s", err.Error()))
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "Failed to retrieve all orders"}}
		} else {
			logrus.Info("Successfully retrieved all orders")
			response = utils.JsonResponse{Data: orders}
		}
	} else {
		id, err := strconv.Atoi(idParam)
		if err != nil {
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "'id' parameter could not be converted to an integer" + err.Error()}}
		}

		if response.Error == nil {
			logrus.Info("Retrieving order with id " + strconv.Itoa(id) + "...")
			var order *models.FruitOrder
			order, err := h.repo.GetByID(int64(id))
			if err != nil {
				logrus.Error(fmt.Sprintf("Error retrieving order with id = %d: %s", id, err.Error()))
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "Failed to retrieve order with id = " + strconv.Itoa(id)}}
			} else {
				logrus.Info("Successfully retrieved order with id = " + strconv.Itoa(id))
				response = utils.JsonResponse{Data: []*models.FruitOrder{order}}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else {
		utils.WriteJSONResponse(w, http.StatusOK, response)
	}
}

// UpdateOrder updates an existing order in the repo based on the supplied JSON request, and returns a status message in JSON to the user.
func (h *Order) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodPut, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf("Update API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}

	var response = utils.JsonResponse{}
	var order models.FruitOrder
	err := utils.DecodeJSONBody(w, r, &order, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if err != nil {
		var request *utils.MalformedRequestError
		if errors.As(err, &request) {
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: request.Status, Message: request.Message}}
		} else {
			msg := "Failed to decode JSON body: " + err.Error()
			logrus.Error(msg)
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: msg}}
		}
	}

	if response.Error == nil {
		isOrderValid, orderValidationError := models.ValidateFruitOrder(&order)
		if !isOrderValid {
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: orderValidationError.Error()}}
		} else {
			logrus.Info(fmt.Sprintf("Trying to update FruitOrder with id = %d to %+v", order.ID, order))
			_, err = h.repo.Update(&order)
			if err != nil {
				logrus.Error(fmt.Sprintf("Error updating order with id = %d: %+ve: %s", order.ID, order, err.Error()))
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create new order."}}
			} else {
				logrus.Info(fmt.Sprintf("Successfully updated order with id = %d: %+v", order.ID, order))
				response = utils.JsonResponse{Data: []*models.FruitOrder{&order}, Id: strconv.Itoa(int(order.ID))}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else {
		utils.WriteJSONResponse(w, http.StatusOK, response)
	}
}

// DeleteOrder deletes an existing order from the repo based on the supplied http request, and returns a status message in JSON to the user.
func (h *Order) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodDelete, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf("Register API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}

	var response = utils.JsonResponse{}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "'id' parameter must be specified"}}
	}
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "'id' parameter could not be converted to an integer" + err.Error()}}
	}

	if response.Error == nil {
		logrus.Info("Deleting order with id " + strconv.Itoa(id) + "...")
		_, err := h.repo.Delete(int64(id))
		if err != nil {
			logrus.Error(fmt.Sprintf("Error deleting order with id = %d: %s", id, err.Error()))
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete order with id = " + strconv.Itoa(id)}}
		} else {
			logrus.Info(fmt.Sprintf("Successfully deleted order with id = %d.", id))
			//response = utils.JsonResponse{Id: strconv.Itoa(id)}
		}

	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else {
		utils.WriteJSONResponse(w, http.StatusNoContent, response)
	}
}
