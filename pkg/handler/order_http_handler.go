package handler

import (
	"errors"
	"strings"

	"github.com/tragicpixel/fruitbar/pkg/driver"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	itemsrepo "github.com/tragicpixel/fruitbar/pkg/repository/item"
	jwtrepo "github.com/tragicpixel/fruitbar/pkg/repository/jwt"
	orderrepo "github.com/tragicpixel/fruitbar/pkg/repository/order"
	productsrepo "github.com/tragicpixel/fruitbar/pkg/repository/product"
	"github.com/tragicpixel/fruitbar/pkg/utils"
	jwtutils "github.com/tragicpixel/fruitbar/pkg/utils/jwt"

	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Order represents an http handler for performing operations on a repository of orders.
type Order struct {
	repo         repository.Order
	productsRepo repository.Product
	itemsRepo    repository.Item
	jwtRepo      repository.Jwt
}

// NewOrderHandler creates a new http handler for performing operations on a repository of orders.
func NewOrderHandler(db *driver.DB) *Order {
	return &Order{
		repo:         orderrepo.NewPostgresOrderRepo(db.Postgres),
		productsRepo: productsrepo.NewPostgresProductRepo(db.Postgres),
		itemsRepo:    itemsrepo.NewPostgresItemRepo(db.Postgres),
		jwtRepo:      jwtrepo.NewJWTRepository(),
	}
}

const (
	validationFailedMsg        = "validation failed: "
	internalServerErrMsg       = "Internal server error. Please contact your system administrator."
	forbiddenReadOrderErrMsg   = forbiddenErrMsg + "read this Order."
	forbiddenUpdateOrderErrMsg = forbiddenErrMsg + "update this Order."
	forbiddenDeleteOrderErrMsg = forbiddenErrMsg + "delete this Order."

	idParam     = "id"
	fieldsParam = "fields"

	readPageMaxLimit = 10
)

// CreateOrder creates a new order in the repo based on the supplied HTTP request and sends a response in JSON to the user based on success or failure.
// Requires at least 1 fruit to be purchased, paymentInfo.cash must be true or paymentInfo.cardInfo must be filled out and valid.
// Any supplied values for subtotal, tax, and total, will be overwritten.
func (h *Order) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Decode the new order from JSON format TODO(tragicpixel): Make this block its own function
	var response = utils.JsonResponse{}
	var order models.Order
	response = *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &order, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}

	// Validate the new order, and each of its items
	_, err := models.ValidateNewOrder(&order)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, validationFailedMsg+err.Error())
		return
	}
	_, err = h.validateItems(order.Items)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, validationFailedMsg+err.Error())
		return
	}

	// Calculate the subtotal, tax, and total fields on the order
	subtotal, err := h.calculateOrderSubtotal(&order)
	if err != nil {
		logMsg := "Failed to calculate new order subtotal: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	order.Subtotal = subtotal
	order.Tax = order.Subtotal * order.TaxRate
	order.Total = order.Subtotal + order.Tax

	// Create the new order
	logrus.Info("Inserting new order...")
	orderID, itemIds, err := h.repo.Create(&order)
	if err != nil {
		logMsg := fmt.Sprintf("Error inserting order %+v into database: %s", order, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	for _, id := range itemIds {
		update := models.Item{OrderID: orderID}
		update.ID = id
		_, err := h.itemsRepo.Update(&update, []string{"orderid", "id"})
		if err != nil {
			logMsg := fmt.Sprintf("Error updating item (id: %d) for order (id: %d): %s", id, order.ID, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
	}
	logrus.Info(fmt.Sprintf("Created new order (id: %d): %+v", orderID, order))
	//order.ID = orderID
	response = utils.JsonResponse{Data: []*models.Order{&order}}
	utils.WriteJSONResponse(w, http.StatusCreated, response)
}

// GetOrders retrieves an existing order in the repo based on the supplied ID query parameter and returns a response in JSON containing either the order encoded in JSON or an error message.
func (h *Order) GetOrders(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has(idParam) {
		h.getSingleOrder(w, r)
	} else {
		h.getOrdersPage(w, r)
	}
}

func (h *Order) getSingleOrder(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	logrus.Info(fmt.Sprintf("Reading order with id %d...", id))
	var order *models.Order
	order, err = h.repo.GetByID(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error reading order (id: %d) %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Successfully retrieved order with id = %d", id))

	requestor, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := "Authorization failed: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Authorization Failed", logMsg)
		return
	}
	// Customers can only read orders with their own IDs
	if requestor.UserRole == models.GetCustomerRoleId() && order.OwnerID != requestor.UserID {
		http.Error(w, forbiddenReadOrderErrMsg, http.StatusForbidden)
		return
	}

	response := utils.JsonResponse{Data: []*models.Order{order}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (h *Order) getOrdersPage(w http.ResponseWriter, r *http.Request) {
	var seek *utils.PageSeekOptions
	seek, err := utils.GetPageSeekOptions(r, readPageMaxLimit)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	logrus.Info(fmt.Sprintf("Reading %d orders (max %d)...", seek.RecordLimit, readPageMaxLimit))
	var orders []*models.Order
	orders, err = h.repo.Fetch(seek)
	if err != nil {
		logMsg := fmt.Sprintf("Error reading orders: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}

	requestor, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := "Authorization failed: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Authorization Failed", logMsg)
		return
	}
	var pruned []*models.Order
	switch requestor.UserRole {
	case models.GetCustomerRoleId():
		// Customers can only read: orders owned by their user ID
		for _, order := range orders {
			if order.OwnerID == requestor.UserID {
				pruned = append(pruned, order)
			}
		}
	default:
		pruned = orders
	}
	orders = pruned

	count, err := h.repo.Count(seek)
	if err != nil {
		logMsg := fmt.Sprintf("Error counting orders: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	startID, endID := uint(0), uint(0)
	if len(orders) > 0 {
		startID = orders[0].ID
		endID = orders[len(orders)-1].ID
	}
	rangeStr := fmt.Sprintf("orders%d-%d/%d", startID, endID, count)
	w.Header().Set("Content-Range", rangeStr)
	logrus.Info(fmt.Sprintf("Read %d orders", len(orders)))
	response := utils.JsonResponse{Data: orders}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// UpdateOrder updates an existing order in the repo based on the supplied JSON request, and returns a status message in JSON to the user.
func (h *Order) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	var response = utils.JsonResponse{}
	var order models.Order
	response = *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &order, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}

	_, err := h.validateItems(order.Items)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Items "+validationFailedMsg+err.Error())
		return
	}

	requestor, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := "Authorization failed: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Authorization Failed", logMsg)
		return
	}
	// Customers can only update their own orders
	if requestor.UserRole == models.GetCustomerRoleId() && order.OwnerID != requestor.UserID {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenUpdateOrderErrMsg)
		return
	}

	if r.URL.Query().Has(fieldsParam) {
		h.partiallyUpdateOrder(w, r, order)
	} else {
		h.fullyUpdateOrder(w, r, order)
	}
}

// partiallyUpdateOrder updates only the specified fields (via http query parameter) of the supplied order.
// Sends a response in json to the supplied http ResponseWriter.
func (h *Order) partiallyUpdateOrder(w http.ResponseWriter, r *http.Request, order models.Order) {
	fieldsStr := r.URL.Query().Get(fieldsParam)
	fields := strings.Split(fieldsStr, ",")

	_, err := models.ValidateOrderUpdate(&order, fields)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Order "+validationFailedMsg+err.Error())
		return
	}

	logrus.Info(fmt.Sprintf("Updating order (id: %d) fields (%s) to %+v", order.ID, fieldsStr, order))
	updated, err := h.repo.Update(&order, fields)
	if err != nil {
		logMsg := fmt.Sprintf("Error partially updating order (id: %d) fields (%s) to %+v: %s", order.ID, fieldsStr, order, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Partially updated order (id: %d) fields (%s): %+v", order.ID, fieldsStr, updated))
	currentItems, err := h.itemsRepo.GetByOrderID(order.ID)
	if err != nil {
		logMsg := fmt.Sprintf("Failed to retrieve existing items: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	for _, item := range order.Items {
		match := false
		for _, existingItem := range currentItems {
			if item.ProductID == existingItem.ProductID {
				match = true
				item.ID = existingItem.ID
				logrus.Info(fmt.Sprintf("Updating item (id: %d) to %+v", item.ID, item))
				_, err := h.itemsRepo.Update(&item, []string{})
				if err != nil {
					logMsg := fmt.Sprintf("Error updating item (id: %d): %s", item.ID, err.Error())
					utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
					return
				}
				break // assumption: only one item exists with the given product id
			}
		}
		if !match {
			logrus.Info(fmt.Sprintf("Creating new item: %+v", item))
			item.OrderID = order.ID
			id, err := h.itemsRepo.Create(&item)
			if err != nil {
				logMsg := fmt.Sprintf("Error creating item: %s", err.Error())
				utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
				return
			}
			item.ID = id
		}
	}
	logrus.Info(fmt.Sprintf("Partially updated order (id: %d) fields: %s", order.ID, fieldsStr))
	response := utils.JsonResponse{Data: []*models.Order{&order}, Id: strconv.Itoa(int(order.ID))}
	utils.WriteJSONResponse(w, http.StatusOK, response)
	return
}

// fullyUpdateOrder updates all of the fields of the supplied order.
// Sends a response in json to the supplied http ResponseWriter.
func (h *Order) fullyUpdateOrder(w http.ResponseWriter, r *http.Request, order models.Order) {
	_, err := models.ValidateOrder(&order)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Order "+validationFailedMsg+err.Error())
		return
	}

	logrus.Info(fmt.Sprintf("Updating order (id: %d) to %+v", order.ID, order))
	updated, err := h.repo.Update(&order, []string{})
	if err != nil {
		logMsg := fmt.Sprintf("Error updating order (id: %d): %s", order.ID, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Updated order (id: %d) to %+v", order.ID, updated))

	currentItems, err := h.itemsRepo.GetByOrderID(order.ID)
	if err != nil {
		logMsg := fmt.Sprintf("Failed to retrieve existing items: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	for _, item := range currentItems {
		logrus.Info(fmt.Sprintf("Deleting existing item (id: %d) from order (id: %d)...", item.ID, order.ID))
		err := h.itemsRepo.Delete(item.ID)
		if err != nil {
			logMsg := fmt.Sprintf("Error deleting existing item: %s", err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
	}
	for _, item := range order.Items {
		logrus.Info(fmt.Sprintf("Creating new item for order (id: %d)...", order.ID))
		id, err := h.itemsRepo.Create(&item)
		if err != nil {
			logMsg := fmt.Sprintf("couldn't create an item for a full order update: %s", err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
		item.ID = id
	}
	logrus.Info(fmt.Sprintf("Fully updated order (id: %d)", order.ID))
	response := utils.JsonResponse{Data: []*models.Order{&order}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// DeleteOrder deletes an existing order from the repo based on the supplied http request and writes a response over http.
func (h *Order) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.repo.GetByID(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error reading order for proposed deletion...")
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	requestor, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := "Authorization failed: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Authorization Failed", logMsg)
		return
	}
	// Customers can only delete their own orders
	if requestor.UserRole == models.GetCustomerRoleId() && order.OwnerID != requestor.UserID {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenDeleteOrderErrMsg)
		return
	}

	existingItems, err := h.itemsRepo.GetByOrderID(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error reading existing items for order (id: %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	for _, item := range existingItems {
		logrus.Info(fmt.Sprintf("Deleting existing item (id: %d) from order (id: %d)", item.ID, id))
		err := h.itemsRepo.Delete(item.ID)
		if err != nil {
			logMsg := fmt.Sprintf("Error deleting existing item (id: %d): %s", item.ID, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
	}
	logrus.Info(fmt.Sprintf("Deleted all items for order (id: %d)", id))

	logrus.Info(fmt.Sprintf("Deleting order (id: %d)..., ", id))
	err = h.repo.Delete(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error deleting order (id %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	successMsg := fmt.Sprintf("Deleted order (id: %d)", id)
	logrus.Info(successMsg)
	response := utils.JsonResponse{Data: successMsg}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

func (h *Order) validateItemsProductID(items []models.Item) (bool, error) {
	ids := make(map[uint]bool, len(items))
	for _, item := range items {
		exists, err := h.productsRepo.Exists(item.ProductID)
		if err != nil {
			return false, errors.New("failed to validate product id: " + err.Error())
		}
		if !exists {
			return false, fmt.Errorf("product ID %d does not exist in the repo", item.ProductID)
		}
		_, idAlreadyExists := ids[item.ProductID]
		if idAlreadyExists {
			return false, errors.New("item list contains duplicate product ID: " + strconv.Itoa(int(item.ProductID)))
		}
		ids[item.ProductID] = true
	}
	return true, nil
}

func (h *Order) validateItemsQuantity(items []models.Item) (bool, error) {
	for _, item := range items {
		if item.Quantity <= 0 {
			return false, errors.New("an item's quantity must be greater than zero")
		}
	}
	return true, nil
}

func (h *Order) validateItems(items []models.Item) (bool, error) {
	_, err := h.validateItemsProductID(items)
	if err != nil {
		return false, err
	}
	_, err = h.validateItemsQuantity(items)
	if err != nil {
		return false, err
	}
	return true, nil
}

// calculateOrderSubtotal returns the calculated subtotal based on the supplied order.
func (h *Order) calculateOrderSubtotal(order *models.Order) (float64, error) {
	subtotal := float64(0)
	for _, item := range order.Items {
		product, err := h.productsRepo.GetByID(item.ProductID)
		if err != nil {
			return -1, err
		}
		subtotal += float64(item.Quantity) * product.Price
	}
	return subtotal, nil
}
