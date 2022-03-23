package handler

import (
	"errors"
	"strings"

	"github.com/tragicpixel/fruitbar/pkg/driver"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/models/roles"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	itemsrepo "github.com/tragicpixel/fruitbar/pkg/repository/item"
	jwtrepo "github.com/tragicpixel/fruitbar/pkg/repository/jwt"
	orderrepo "github.com/tragicpixel/fruitbar/pkg/repository/order"
	productsrepo "github.com/tragicpixel/fruitbar/pkg/repository/product"
	"github.com/tragicpixel/fruitbar/pkg/utils"
	jwtutils "github.com/tragicpixel/fruitbar/pkg/utils/jwt"
	"gorm.io/gorm"

	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Order represents a handler for performing operations on orders via HTTP.
type Order struct {
	repo         repository.Order
	productsRepo repository.Product
	itemsRepo    repository.Item
	jwtRepo      repository.Jwt
}

// NewOrderHandler creates and initializes a new handler for performing operations on orders via HTTP.
func NewOrderHandler(db *driver.DB) *Order {
	return &Order{
		repo:         orderrepo.NewPostgresOrderRepo(db.Postgres),
		productsRepo: productsrepo.NewPostgresProductRepo(db.Postgres),
		itemsRepo:    itemsrepo.NewPostgresItemRepo(db.Postgres),
		jwtRepo:      jwtrepo.NewJWTRepository(),
	}
}

// CreateOrder creates a new order based on the supplied HTTP request and sends a response in JSON containing the newly created order to the supplied http response writer.
// If there is a permission error, an HTTP error will be sent.
func (h *Order) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var response = utils.JsonResponse{}
	var order models.Order
	response = *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &order, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}

	if !h.clientHasCreatePermsForOrder(w, r, order) {
		return
	}

	_, err := models.ValidateNewOrder(&order)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Order "+validationFailedErrMsgPrefix+err.Error())
		return
	}
	if err := h.itemsAreValid(order.Items); err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Items "+validationFailedErrMsgPrefix+err.Error())
		return
	}

	subtotal, err := h.calculateOrderSubtotal(&order)
	if err != nil {
		logMsg := "Failed to calculate new order subtotal: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	order.Subtotal = subtotal
	order.Tax = order.Subtotal * order.TaxRate
	order.Total = order.Subtotal + order.Tax

	logrus.Info("Inserting new order...")
	createdID, itemIds, err := h.repo.Create(&order)
	if err != nil {
		logMsg := fmt.Sprintf("Error inserting order %+v into database: %s", order, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	// Update all newly created items (gorm creates them) with the order's ID (didn't know the order ID until created)
	// TODO: There is probably a way to create a key constraint in gorm so it does this automatically
	for _, id := range itemIds {
		update := models.Item{OrderID: createdID}
		update.ID = id
		logrus.Info(fmt.Sprintf("Updating item (id: %d) for order (id: %d)", id, order.ID))
		_, err := h.itemsRepo.Update(&update, []string{"orderid", "id"})
		if err != nil {
			logMsg := fmt.Sprintf("Error updating item (id: %d) for order (id: %d): %s", id, order.ID, err.Error())
			utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
			return
		}
	}
	logrus.Info(fmt.Sprintf("Created new order (id: %d): %+v", createdID, order))
	response = utils.JsonResponse{Data: []*models.Order{&order}}
	utils.WriteJSONResponse(w, http.StatusCreated, response)
}

// GetOrders sends a response to the supplied http response writer containing the requested order(s), based on the supplied http request.
func (h *Order) GetOrders(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has(idParam) {
		h.getSingleOrder(w, r)
	} else {
		h.getOrdersPage(w, r)
	}
}

// UpdateOrder updates an existing order based on the supplied http request and sends a response in JSON containing the updated order to the supplied http response writer.
func (h *Order) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	var response = utils.JsonResponse{}
	var order models.Order
	response = *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &order, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error != nil {
		utils.WriteJSONErrorResponse(w, response.Error.Code, response.Error.Message)
		return
	}

	if !h.clientHasUpdatePermsForOrder(w, r, order) {
		return
	}

	if err := h.itemsAreValid(order.Items); err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Items "+validationFailedErrMsgPrefix+err.Error())
		return
	}

	if r.URL.Query().Has(fieldsParam) {
		h.partiallyUpdateOrder(w, r, order)
	} else {
		h.fullyUpdateOrder(w, r, order)
	}
}

// DeleteOrder deletes an existing order and all of its child items based on the supplied http request and sends a status code to the supplied http response writer.
func (h *Order) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	logrus.Info(fmt.Sprintf("Reading order (id: %d) for proposed deletion...", id))
	order, err := h.repo.GetByID(id)
	if err != nil {
		logMsg := "Error reading order for proposed deletion: " + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}

	if !h.clientHasDeletePermsForOrder(w, r, order) {
		return
	}

	logrus.Info(fmt.Sprintf("Reading items for order (id: %d) for proposed deletion...", id))
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
	logrus.Info(fmt.Sprintf("Deleted order (id: %d)", id))

	w.WriteHeader(http.StatusNoContent)
}

// getSingleOrder sends a response to the supplied http response writer containing the requested order, based on the supplied http request.
// If read access to a specific order is forbidden, an error will be sent.
func (h *Order) getSingleOrder(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetQueryParamAsUint(r, idParam)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	logrus.Info(fmt.Sprintf("Selecting order with id %d...", id))
	var order *models.Order
	order, err = h.repo.GetByID(id)
	if err != nil {
		logMsg := fmt.Sprintf("Error selecting order (id: %d): %s", id, err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	logrus.Info(fmt.Sprintf("Successfully selected order with id = %d", id))

	if !h.clientHasReadPermsForOrder(w, r, order) {
		return
	}

	response := utils.JsonResponse{Data: []*models.Order{order}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// getOrdersPage sends a response to the supplied http response writer containing the requested page of orders, based on the supplied http request.
// If read access to a specific order is forbidden, it won't be included in the response and there will be no error message. (fail silently)
func (h *Order) getOrdersPage(w http.ResponseWriter, r *http.Request) {
	var seek *utils.PageSeekOptions
	seek, err := utils.GetPageSeekOptions(r, readPageMaxLimit)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	logrus.Info(fmt.Sprintf("Selecting %d orders (max %d)...", seek.RecordLimit, readOrdersMaxRecordLimit))
	var orders []*models.Order
	orders, err = h.repo.Fetch(seek)
	if err != nil {
		logMsg := fmt.Sprintf("Error selecting orders: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}

	if orders = h.getOrdersReadableByClient(w, r, orders); orders == nil {
		return
	}

	rangeStr := h.getOrdersRangeStr(w, seek, orders)
	w.Header().Set("Content-Range", rangeStr)
	logrus.Info(fmt.Sprintf("Read %d orders", len(orders)))
	response := utils.JsonResponse{Data: orders}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// partiallyUpdateOrder updates only the specified fields (from the supplied http request) of the order and sends a response in JSON containing the newly updated order.
// Assumes permission check has already been performed.
func (h *Order) partiallyUpdateOrder(w http.ResponseWriter, r *http.Request, order models.Order) {
	fieldsStr := r.URL.Query().Get(fieldsParam)
	fields := strings.Split(fieldsStr, ",")

	_, err := models.ValidateOrderUpdate(&order, fields)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Order "+validationFailedErrMsgPrefix+err.Error())
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

	existingItems, err := h.itemsRepo.GetByOrderID(order.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logMsg := fmt.Sprintf("Failed to select existing items: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return
	}
	for _, item := range order.Items {
		match := false
		for _, existingItem := range existingItems {
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
				break // data assumption: product IDs on items are unique (there is a maximum of 1 item with any given product ID)
			}
		}
		if !match {
			logrus.Info(fmt.Sprintf("Inserting new item: %+v", item))
			item.OrderID = order.ID
			id, err := h.itemsRepo.Create(&item)
			if err != nil {
				logMsg := fmt.Sprintf("Error inserting item: %s", err.Error())
				utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
				return
			}
			item.ID = id
		}
	}
	logrus.Info(fmt.Sprintf("Updated order's items (id: %d) due to partial update", order.ID))
	response := utils.JsonResponse{Data: []*models.Order{&order}}
	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// fullyUpdateOrder updates all the fields of the order (based on the supplied http request) and sends a response in JSON containing the newly updated order.
// Assumes permission check has already been performed.
// Note: All items for the given order will be deleted, and the items included in the updated order will be created.
func (h *Order) fullyUpdateOrder(w http.ResponseWriter, r *http.Request, order models.Order) {
	_, err := models.ValidateOrder(&order)
	if err != nil {
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, "Order "+validationFailedErrMsgPrefix+err.Error())
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

	logrus.Info(fmt.Sprintf("Selecting existing items for order (id: %d", order.ID))
	currentItems, err := h.itemsRepo.GetByOrderID(order.ID)
	if err != nil {
		logMsg := fmt.Sprintf("Failed to select existing items: %s", err.Error())
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
		logrus.Info(fmt.Sprintf("Inserting new item for order (id: %d)...", order.ID))
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

// itemsAreValid validates whether the supplied items are valid.
func (h *Order) itemsAreValid(items []models.Item) error {
	_, err := h.validateProductIDs(items)
	if err != nil {
		return err
	}
	_, err = h.quantityIsValid(items)
	if err != nil {
		return err
	}
	return nil
}

// validateProductIDs checks that the product ID values in the supplied set of items, correspond to products that actually exist.
func (h *Order) validateProductIDs(items []models.Item) (bool, error) {
	ids := make(map[uint]bool, len(items))
	for _, item := range items {
		// TODO: Rewrite this so that only one database call is made -> modify Exists() to take var args and send all the IDs at once
		logrus.Info(fmt.Sprintf("Checking if a product with ID = %d exists", item.ProductID))
		exists, err := h.productsRepo.Exists(item.ProductID)
		if err != nil {
			return false, errors.New("failed to validate product id: " + err.Error())
		}
		if !exists {
			return false, fmt.Errorf("product ID %d does not exist in the repo", item.ProductID)
		}
		_, idAlreadyExists := ids[item.ProductID]
		if idAlreadyExists {
			return false, fmt.Errorf("item list contains duplicate product ID: %d", item.ProductID)
		}
		ids[item.ProductID] = true
	}
	return true, nil
}

// quantityIsValid validates whether the quantity values of the supplied items are all valid.
func (h *Order) quantityIsValid(items []models.Item) (bool, error) {
	for _, item := range items {
		if item.Quantity <= 0 {
			return false, fmt.Errorf("an item's quantity must be greater than zero. got %d", item.Quantity)
		}
	}
	return true, nil
}

// calculateOrderSubtotal returns the calculated subtotal based on the supplied order.
func (h *Order) calculateOrderSubtotal(order *models.Order) (float64, error) {
	subtotal := float64(0)
	for _, item := range order.Items {
		logrus.Info(fmt.Sprintf("Selecting product (id: %d) to get price...", item.ProductID))
		product, err := h.productsRepo.GetByID(item.ProductID)
		if err != nil {
			return -1, err
		}
		subtotal += float64(item.Quantity) * product.Price
	}
	return subtotal, nil
}

// getClientAuthInfo returns the authorization information about the client based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *Order) getClientAuthInfo(w http.ResponseWriter, r *http.Request) *models.JwtClaim {
	client, err := jwtutils.GetTokenClaims(r, h.jwtRepo)
	if err != nil {
		logMsg := unauthorizedErrMsgPrefix + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusBadRequest, unauthorizedErrMsg, logMsg)
		return nil
	}
	_, err = roles.IsValid(client.UserRole)
	if err != nil {
		logMsg := unauthorizedErrMsgPrefix + err.Error()
		utils.WriteJSONErrorResponse(w, http.StatusUnauthorized, unauthorizedErrMsg, logMsg)
		return nil
	}
	return client
}

// clientHasCreatePermsForOrder checks whether the client has permissions to create the supplied order, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *Order) clientHasCreatePermsForOrder(w http.ResponseWriter, r *http.Request, order models.Order) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	// Customers can only create orders owned by themselves
	if client.UserRole == roles.Customer && order.OwnerID != client.UserID {
		http.Error(w, unauthorizedErrMsg, http.StatusUnauthorized)
		return false
	}
	return true
}

// clientHasReadPermsForOrder checks whether the client has permissions to read the supplied order, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *Order) clientHasReadPermsForOrder(w http.ResponseWriter, r *http.Request, order *models.Order) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	// Customers can only read orders with their own IDs
	if client.UserRole == roles.Customer && order.OwnerID != client.UserID {
		http.Error(w, forbiddenReadOrderErrMsg, http.StatusForbidden)
		return false
	}
	return true
}

// clientHasCreatePermsForOrder prunes the supplied orders down to only the orders the client has permission to read.
func (h *Order) getOrdersReadableByClient(w http.ResponseWriter, r *http.Request, orders []*models.Order) []*models.Order {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return nil
	}
	var pruned []*models.Order
	switch client.UserRole {
	case roles.Customer:
		// Customers can only read orders owned by their user ID
		for _, order := range orders {
			if order.OwnerID == client.UserID {
				pruned = append(pruned, order)
			}
		}
	default:
		pruned = orders
	}
	return pruned
}

// clientHasUpdatePermsForOrder checks whether the client has permissions to update the supplied order, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *Order) clientHasUpdatePermsForOrder(w http.ResponseWriter, r *http.Request, order models.Order) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	// Customers can only update their own orders
	if client.UserRole == roles.Customer && order.OwnerID != client.UserID {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenUpdateOrderErrMsg)
		return false
	}
	return true
}

// clientHasDeletePermsForOrder checks whether the client has permissions to delete the supplied order, based on the supplied http request.
// Writes a response on the supplied http response writer if there is an error.
func (h *Order) clientHasDeletePermsForOrder(w http.ResponseWriter, r *http.Request, order *models.Order) bool {
	client := h.getClientAuthInfo(w, r)
	if client == nil {
		return false
	}
	// Customers can only delete their own orders
	if client.UserRole == roles.Customer && order.OwnerID != client.UserID {
		utils.WriteJSONErrorResponse(w, http.StatusForbidden, forbiddenDeleteOrderErrMsg)
		return false
	}
	return true
}

// getUsersRangeStr returns a string representation of the range of the supplied orders.
func (h *Order) getOrdersRangeStr(w http.ResponseWriter, seek *utils.PageSeekOptions, orders []*models.Order) string {
	logrus.Info("Counting orders...")
	count, err := h.repo.Count(seek)
	if err != nil {
		logMsg := fmt.Sprintf("Error counting orders: %s", err.Error())
		utils.WriteJSONErrorResponse(w, http.StatusInternalServerError, internalServerErrMsg, logMsg)
		return ""
	}
	startID, endID := uint(0), uint(0)
	if len(orders) > 0 {
		startID = orders[0].ID
		endID = orders[len(orders)-1].ID
	}
	rangeStr := fmt.Sprintf("orders%d-%d/%d", startID, endID, count)
	return rangeStr
}
