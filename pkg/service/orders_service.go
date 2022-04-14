package service

import (
	"github.com/tragicpixel/fruitbar/pkg/driver"
	pgdriver "github.com/tragicpixel/fruitbar/pkg/driver/postgres"
	"github.com/tragicpixel/fruitbar/pkg/handler"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/utils/cors"
	"github.com/tragicpixel/fruitbar/pkg/utils/json"
	"github.com/tragicpixel/fruitbar/pkg/utils/log"

	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// OrdersService holds all the pieces necessary to run the data entry service for the fruitbar application.
type OrdersService struct {
	Router          *mux.Router
	Handler         *handler.Order
	UserHandler     *handler.User
	DB              *driver.DB
	Port            int
	SalesTaxPercent float64
}

type OrdersServiceConfig struct {
	DatabaseConnection *pgdriver.PostgresConnectionConfig
	Port               int
	SalesTaxPercent    float64
}

const (
	ordersAPIBaseRoute   = "/orders/"
	ordersCreateAPIRoute = ordersAPIBaseRoute
	ordersReadAPIRoute   = ordersAPIBaseRoute
	ordersUpdateAPIRoute = ordersAPIBaseRoute
	ordersDeleteAPIRoute = ordersAPIBaseRoute
	ordersHealthAPIRoute = ordersAPIBaseRoute + "health"
)

func (s *OrdersService) getOrdersEndpointOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Orders Options",
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}
}
func (s *OrdersService) getCreateAPICORSOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Create Order",
		AllowedMethods: []string{http.MethodPost},
	}
}
func (s *OrdersService) getReadAPICORSOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Read Order",
		AllowedMethods: []string{http.MethodGet},
	}
}
func (s *OrdersService) getUpdateAPICORSOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Update Order",
		AllowedMethods: []string{http.MethodPut},
	}
}
func (s *OrdersService) getDeleteAPICORSOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Delete Order",
		AllowedMethods: []string{http.MethodDelete},
	}
}
func (s *OrdersService) getHealthCheckAPICORSOptions() cors.Options {
	return cors.Options{
		AllowedURL:     UI_URL,
		APIName:        "Health Check",
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
	}
}

func (s *OrdersService) getCreateAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getCreateAPICORSOptions(), s.UserHandler.IsAuthorized(s.Handler.CreateOrder))
}
func (s *OrdersService) getReadAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getReadAPICORSOptions(), s.UserHandler.IsAuthorized(s.Handler.GetOrders))
}
func (s *OrdersService) getUpdateAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getUpdateAPICORSOptions(), s.UserHandler.IsAuthorized(s.Handler.UpdateOrder))
}
func (s *OrdersService) getDeleteAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getDeleteAPICORSOptions(), s.UserHandler.IsAuthorized(s.Handler.DeleteOrder))
}
func (s *OrdersService) getHealthCheckAPIHandler() func(http.ResponseWriter, *http.Request) {
	return cors.SendPreflightHeaders(s.getHealthCheckAPICORSOptions(), s.CheckHealth)
}

// NewOrdersService creates a new instance of a data entry service.
// Returns nil on error.
func NewOrdersService(config *OrdersServiceConfig) (*OrdersService, error) {
	s := OrdersService{}

	db, err := pgdriver.OpenConnection(config.DatabaseConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the orders service database: %s", err.Error())
	}

	s.DB = db
	err = setupOrdersServiceDB(s.DB, true)
	if err != nil {
		return nil, fmt.Errorf("failed to set up the orders service database: %s", err.Error())
	}

	s.Handler = handler.NewOrderHandler(db)
	s.UserHandler = handler.NewUserHandler(db)
	s.Router = s.NewOrdersServiceRouter(db)
	s.Port = config.Port
	s.SalesTaxPercent = config.SalesTaxPercent

	return &s, nil
}

// NewOrdersServiceRouter creates and returns a new http router for the data entry service.
func (s *OrdersService) NewOrdersServiceRouter(db *driver.DB) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc(productsAPIBaseRoute, cors.SendPreflightHeaders(s.getOrdersEndpointOptions(), nil)).Methods(http.MethodOptions)
	// swagger:operation POST /orders/ orders createOrder
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
	r.HandleFunc(ordersCreateAPIRoute, s.getCreateAPIHandler()).Methods(s.getCreateAPICORSOptions().AllowedMethods...)
	// swagger:operation GET /orders/ orders getOrder
	//
	// Get an order by ID, or a paginated listing of all orders.
	//
	// ---
	// parameters:
	// - name: id
	//   in: query
	//   description: id of order to retrieve.
	//   required: false
	//   schema:
	//     type: int
	// security:
	// - bearer: []
	// responses:
	//   '200':
	//     description: Successfully retrieved an order.
	//     "$ref": "#/responses/jsonResponse"
	r.HandleFunc(ordersReadAPIRoute, s.getReadAPIHandler()).Methods(s.getReadAPICORSOptions().AllowedMethods...)
	// swagger:operation PUT /orders/ orders updateOrder
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
	r.HandleFunc(ordersUpdateAPIRoute, s.getUpdateAPIHandler()).Methods(s.getUpdateAPICORSOptions().AllowedMethods...)
	// swagger:operation DELETE /orders/ orders deleteOrder
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
	r.HandleFunc(ordersDeleteAPIRoute, s.getDeleteAPIHandler()).Methods(s.getDeleteAPICORSOptions().AllowedMethods...)
	// swagger:operation GET /orders/health orders checkHealth
	//
	// Checks the health of the service.
	//
	// ---
	// responses:
	//   '200':
	//     description: The health check was completed.
	//     "$ref": "#/responses/healthCheckResponse"
	r.HandleFunc(ordersHealthAPIRoute, s.getHealthCheckAPIHandler()).Methods(s.getHealthCheckAPICORSOptions().AllowedMethods...)

	return r
}

// SetupOrdersServiceDB checks that the database schema is ready for the authentication service.
// If init is true, will create the tables if they do not already exist.
func setupOrdersServiceDB(db *driver.DB, init bool) error {
	log.Info("Setting up the orders service database...")
	err := pgdriver.SetupTables(db, &models.Order{}, init)
	if err != nil {
		msg := "failed to set up the Orders model table" + err.Error()
		log.Error(msg)
		return errors.New(msg)
	}
	err = pgdriver.SetupTables(db, &models.Item{}, init)
	if err != nil {
		msg := "failed to set up the Items model table" + err.Error()
		log.Error(msg)
		return errors.New(msg)
	}
	log.Info("Successfully set up the database for the orders service")
	return nil
}

// CheckHealth checks the health of the data entry service and writes a response in JSON to the user.
// Always returns HTTP Status OK, even if the health check fails.
func (s *OrdersService) CheckHealth(w http.ResponseWriter, r *http.Request) {
	var err error
	log.Info("Checking orders service health...")
	db, err := s.DB.Postgres.DB()
	if err != nil {
		log.Error("orders service health check failed: Error getting SQLDB from gorm DB: " + err.Error())
		json.WriteResponse(w, http.StatusOK, map[string]bool{"ok": false})
	} else {
		if err = db.Ping(); err != nil {
			log.Error("orders service health check failed: error pinging the database: " + err.Error())
			json.WriteResponse(w, http.StatusOK, map[string]bool{"ok": false})
		} else {
			log.Info("orders service health check passed")
			json.WriteResponse(w, http.StatusOK, map[string]bool{"ok": true})
		}
	}
}
