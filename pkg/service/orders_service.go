package service

import (
	"github.com/tragicpixel/fruitbar/pkg/driver"
	pgdriver "github.com/tragicpixel/fruitbar/pkg/driver/postgres"
	"github.com/tragicpixel/fruitbar/pkg/handler"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/utils"

	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// OrdersService holds all the pieces necessary to run the data entry service for the fruitbar application.
type OrdersService struct {
	Router      *mux.Router
	Handler     *handler.Order
	UserHandler *handler.User
	DB          *driver.DB
	Port        int
}

// change to /orders ????
const (
	createApiHandlerPath            = "/orders/"
	readApiHandlerPath              = "/orders/"
	updateApiHandlerPath            = "/orders/"
	deleteApiHandlerPath            = "/orders/"
	ordersHealthCheckApiHandlerPath = "/orders/health"
)

func getCreateApiAllowedHttpMethods() []string {
	return []string{http.MethodPost, http.MethodOptions}
}
func getReadApiAllowedHttpMethods() []string {
	return []string{http.MethodGet, http.MethodOptions}
}
func getUpdateApiAllowedHttpMethods() []string {
	return []string{http.MethodPut, http.MethodOptions}
}
func getDeleteApiAllowedHttpMethods() []string {
	return []string{http.MethodDelete, http.MethodOptions}
}
func getHealthCheckApiAllowedHttpMethods() []string {
	return []string{http.MethodGet, http.MethodOptions}
}

// NewOrdersService creates a new instance of a data entry service.
// Returns nil on error.
func NewOrdersService(port int) (*OrdersService, error) {
	s := OrdersService{}

	db, err := pgdriver.OpenConnection("localhost", "5423", "postgres", "fruitbar", "fruitbar")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the orders service database: %s", err.Error())
	}

	s.DB = db
	err = SetupOrdersServiceDB(s.DB, true)
	if err != nil {
		return nil, fmt.Errorf("failed to set up the orders service database: %s", err.Error())
	}

	s.Handler = handler.NewOrderHandler(db)
	s.UserHandler = handler.NewUserHandler(db)
	s.Router = s.NewOrdersServiceRouter(db)
	s.Port = port

	return &s, nil
}

// NewOrdersServiceRouter creates and returns a new http router for the data entry service.
func (s *OrdersService) NewOrdersServiceRouter(db *driver.DB) *mux.Router {
	r := mux.NewRouter()

	// swagger:operation POST /orders orders createOrder
	//
	// Create a new order.
	//
	// ---
	// parameters:
	// - name: order
	//   in: body
	//   description: New order to create. Id, CreatedAt, DeletedAt, UpdatedAt fields will be ignored.
	//   required: true
	//   "$ref": "#/definitions/order"
	// security:
	// - bearer: []
	// responses:
	//   '201':
	//     description: Successfully created an order.
	//     "$ref": "#/responses/jsonResponse"
	//     examples:
	//       application/json: { "ok": 2 }
	//   '400':
	//     description: Invalid request.
	//     "$ref": "#/responses/jsonResponse"
	//   '401':
	//     description: Not authorized.
	//   '403':
	//     description: No authorization header provided.
	//   '405':
	//     description: HTTP method not allowed.
	//   '413':
	//     description: Request body too large.
	//     "$ref": "#/responses/jsonResponse"
	//   '500':
	//     description: Internal server error.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(createApiHandlerPath, s.UserHandler.IsAuthorized(s.Handler.CreateOrder)).Methods(getCreateApiAllowedHttpMethods()...)
	// swagger:operation GET /orders orders getOrder
	//
	// Get an order by ID.
	//
	// ---
	// parameters:
	// - name: id
	//   in: query
	//   description: id of order to retrieve.
	//   required: true
	//   schema:
	//     type: int
	// security:
	// - bearer: []
	// responses:
	//   '200':
	//     description: Successfully retrieved an order.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(readApiHandlerPath, s.UserHandler.IsAuthorized(s.Handler.GetOrders)).Methods(getReadApiAllowedHttpMethods()...)
	// swagger:operation PUT /orders orders updateOrder
	//
	// Update an existing order.
	//
	// ---
	// parameters:
	// - name: order
	//   in: body
	//   description: Order fields to update. CreatedAt, DeletedAt, UpdatedAt fields will be ignored.
	//   required: true
	//   schema:
	//     $ref: "#/definitions/order"
	// security:
	// - bearer: []
	// responses:
	//   '200':
	//     description: Successfully updated an existing order.
	//     "$ref": "#/responses/jsonResponse"
	//   '400':
	//     description: Invalid request.
	//     "$ref": "#/responses/jsonResponse"
	//   '401':
	//     description: Not authorized.
	//   '403':
	//     description: No authorization header provided.
	//   '405':
	//     description: HTTP method not allowed.
	//   '413':
	//     description: Request body too large.
	//     "$ref": "#/responses/jsonResponse"
	//   '500':
	//     description: Internal server error.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(updateApiHandlerPath, s.UserHandler.IsAuthorized(s.Handler.UpdateOrder)).Methods(getUpdateApiAllowedHttpMethods()...)
	// swagger:operation DELETE /orders orders deleteOrder
	//
	// Delete an existing order.
	//
	// ---
	// parameters:
	// - name: id
	//   in: query
	//   description: id of order to delete.
	//   required: true
	//   schema:
	//     type: int
	// security:
	// - bearer: []
	// responses:
	//   '204':
	//     description: Successfully deleted an existing order.
	//   '400':
	//     description: Invalid request.
	//     "$ref": "#/responses/jsonResponse"
	//   '401':
	//     description: Not authorized.
	//   '403':
	//     description: No authorization header provided.
	//   '405':
	//     description: HTTP method not allowed.
	//   '413':
	//     description: Request body too large.
	//     "$ref": "#/responses/jsonResponse"
	//   '500':
	//     description: Internal server error.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(deleteApiHandlerPath, s.UserHandler.IsAuthorized(s.Handler.DeleteOrder)).Methods(getDeleteApiAllowedHttpMethods()...)
	// swagger:operation GET /orders/health orders checkHealth
	//
	// Checks the health of the service.
	//
	// ---
	// responses:
	//   '200':
	//     description: The health check was completed.
	//     "$ref": "#/responses/healthCheckResponse"
	r.HandleFunc(ordersHealthCheckApiHandlerPath, s.CheckHealth).Methods(getHealthCheckApiAllowedHttpMethods()...)

	return r
}

// SetupOrdersServiceDB checks that the database schema is ready for the authentication service.
// If init is true, will create the tables if they do not already exist.
func SetupOrdersServiceDB(db *driver.DB, init bool) error {
	logrus.Info("Setting up the orders service database...")
	err := pgdriver.SetupTables(db, &models.FruitOrder{}, init)
	if err != nil {
		logrus.Error("failed to set up the Orders model table" + err.Error())
		return errors.New("failed to set up the Orders model table: " + err.Error())
	}
	logrus.Info("Successfully set up the database for the orders service")
	return nil
}

// CheckHealth checks the health of the data entry service and writes a response in JSON to the user.
// Always returns HTTP Status OK, even if the health check fails.
func (s *OrdersService) CheckHealth(w http.ResponseWriter, r *http.Request) {
	var err error
	logrus.Info("Checking orders service health...")
	theDatabase, err := s.DB.Postgres.DB()
	if err != nil {
		logrus.Error("orders service health check failed: Error getting SQLDB from gorm DB: " + err.Error())
		utils.WriteJSONResponse(w, http.StatusOK, map[string]bool{"ok": false})
	} else {
		if err = theDatabase.Ping(); err != nil {
			logrus.Error("orders service health check failed: error pinging the database: " + err.Error())
			utils.WriteJSONResponse(w, http.StatusOK, map[string]bool{"ok": false})
		} else {
			logrus.Info("orders service health check passed")
			utils.WriteJSONResponse(w, http.StatusOK, map[string]bool{"ok": true})
		}
	}
}
