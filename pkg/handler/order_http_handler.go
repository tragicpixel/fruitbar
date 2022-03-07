package handler

import (
	"errors"
	"strings"

	"github.com/tragicpixel/fruitbar/pkg/driver"
	"github.com/tragicpixel/fruitbar/pkg/models"
	"github.com/tragicpixel/fruitbar/pkg/repository"
	itemsrepo "github.com/tragicpixel/fruitbar/pkg/repository/item"
	orderrepo "github.com/tragicpixel/fruitbar/pkg/repository/order"
	productsrepo "github.com/tragicpixel/fruitbar/pkg/repository/product"
	"github.com/tragicpixel/fruitbar/pkg/utils"

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
	// need item repo too!
}

// NewOrderHandler creates a new http handler for performing operations on a repository of orders.
func NewOrderHandler(db *driver.DB) *Order {
	return &Order{
		repo:         orderrepo.NewPostgresOrderRepo(db.Postgres), // this is where it is decided which implementation(/database type) of the Order Repo we will use
		productsRepo: productsrepo.NewPostgresProductRepo(db.Postgres),
		itemsRepo:    itemsrepo.NewPostgresItemRepo(db.Postgres),
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
		logrus.Info(fmt.Sprintf("Create Order API: Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}

	var response = utils.JsonResponse{}
	var order models.Order
	response = *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &order, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error == nil { // New order json body was successfully decoded
		// Validate the new order
		newOrderIsValid, newOrderValidationError := models.ValidateNewFruitOrder(&order)
		if !newOrderIsValid {
			msg := "Failed to validate new order: " + newOrderValidationError.Error()
			logrus.Error(msg)
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: msg}}
		} else { // New order is valid
			// Validate the items in the new order
			itemListIsValid, err := h.validateItems(order.Items)
			if !itemListIsValid {
				msg := "Failed to validate new order item list: " + err.Error()
				logrus.Error(msg)
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: msg}}
			} else { // Items are valid

				// Calculate the order subtotal
				subtotal, err := h.calculateOrderSubtotal(&order)
				if err != nil {
					logrus.Error("Failed to calculate new order subtotal: " + err.Error())
					response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Internal server error."}}
				} else { // Subtotal was calculated successfully
					// Create the order
					order.Subtotal = subtotal
					order.Tax = subtotal * 0.05 // TODO: figure out what to do with tax calculation constant
					order.Total = order.Subtotal + order.Tax
					logrus.Info("Calculated order subtotal, tax, total")
					logrus.Info(fmt.Sprintf("Trying to insert new Order: %+v", order))
					createdId, err := h.repo.Create(&order)
					if err != nil {
						logrus.Error(fmt.Sprintf("Error inserting Order %+v into database: %s", order, err.Error()))
						response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create new order."}}
					} else {
						logrus.Info(fmt.Sprintf("Successfully inserted new order: %+v", order))
						response = utils.JsonResponse{Data: []*models.Order{&order}, Id: strconv.Itoa(int(createdId))}
					}
				}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else { // Created order successfully
		utils.WriteJSONResponse(w, http.StatusCreated, response)
	}
}

func (h *Order) getReadApiName() string          { return "Read Order API" }
func (h *Order) getReadOrderIdParamName() string { return "id" }
func (h *Order) getReadPageMaxLimit() int        { return 2 }

// GetOrders retrieves an existing order in the repo based on the supplied ID query parameter and returns a response in JSON containing either the order encoded in JSON or an error message.
func (h *Order) GetOrders(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w, UI_URL)
	allowedMethods := []string{http.MethodGet, http.MethodOptions}
	utils.ValidateHttpRequestMethod(w, r, allowedMethods)
	if r.Method == http.MethodOptions {
		utils.SetCorsPreflightResponseHeaders(&w, allowedMethods)
		logrus.Info(fmt.Sprintf(h.getReadApiName()+": Sent response to CORS preflight request from %s", r.RemoteAddr))
		return
	}

	var response = utils.JsonResponse{}
	if !r.URL.Query().Has(h.getReadOrderIdParamName()) { // Single order query parameter was not set
		// Get page seek options
		var pageSeekOptions utils.PageSeekOptions
		pageSeekOptions, response = utils.GetPageSeekOptions(r, h.getReadPageMaxLimit())
		if response.Error == nil { // Page seek options are valid
			// Read the multiple orders
			logrus.Info("Retrieving " + strconv.Itoa(pageSeekOptions.RecordLimit) + " products (max " + strconv.Itoa(h.getReadPageMaxLimit()) + ")...")
			var orders []*models.Order
			orders, err := h.repo.Fetch(pageSeekOptions)

			if err != nil {
				logrus.Error(fmt.Sprintf("Error retrieving orders list: %s", err.Error()))
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "Failed to retrieve orders list"}}
			} else {
				logrus.Info("Successfully retrieved orders list")
				response = utils.JsonResponse{Data: orders}
			}
		}
	} else { // Single order ID query parameter was set
		// Validate single order ID
		id, err := utils.GetQueryParamAsInt(r, h.getReadOrderIdParamName())
		if err != nil {
			response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}}
		} else { // ID was valid
			// Read the single order
			logrus.Info("Retrieving order with id " + strconv.Itoa(id) + "...")
			var order *models.Order
			order, err := h.repo.GetByID(int64(id))
			if err != nil {
				logrus.Error(fmt.Sprintf("Error retrieving order with id = %d: %s", id, err.Error()))
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: "Failed to retrieve order with id = " + strconv.Itoa(id)}}
			} else {
				logrus.Info("Successfully retrieved order with id = " + strconv.Itoa(id))
				response = utils.JsonResponse{Data: []*models.Order{order}}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else { // Read was successful
		utils.WriteJSONResponse(w, http.StatusOK, response)
	}
}

func (h *Order) getUpdateFieldsParamName() string { return "fields" }

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
	var order models.Order
	response = *utils.DecodeJSONBodyAndGetErrorResponse(w, r, &order, utils.MAX_CREATE_REQUEST_SIZE_IN_BYTES)
	if response.Error == nil { // Order update json successfully decoded
		if r.URL.Query().Has(h.getUpdateFieldsParamName()) { // Partial update of order was requested
			// Validate the partial update's changes
			fieldsStr := r.URL.Query().Get(h.getUpdateFieldsParamName())
			fields := strings.Split(fieldsStr, ",")
			isUpdateValid, err := models.ValidateOrderUpdate(&order, fields)
			if !isUpdateValid {
				response = utils.NewJsonResponseWithError(http.StatusBadRequest, err.Error())
			} else { // Partial update is valid
				// Validate the partial update's items changes
				itemListIsValid, err := h.validateItems(order.Items)
				if !itemListIsValid {
					msg := "Failed to validate partially updated order item list: " + err.Error()
					logrus.Error(msg)
					response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: msg}}
				} else { // Items are valid
					// Partially update the order
					logrus.Info(fmt.Sprintf("Trying to partially update order (id: %d) fields (%s) to %+v", order.ID, fieldsStr, order))
					_, err := h.repo.Update(&order, fields)
					if err != nil {
						logrus.Error(fmt.Sprintf("Error partially updating order (id: %d) fields (%s) to %+v: %s", order.ID, fieldsStr, order, err.Error()))
						response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update order."}}
					} else { // Partial update was successful
						logrus.Info(fmt.Sprintf("Successfully partially updated order (id: %d) fields (%s): %+v", order.ID, fieldsStr, order))
						response = utils.JsonResponse{Data: []*models.Order{&order}, Id: strconv.Itoa(int(order.ID))}

						// Get the set of already existing items for this order
						existingItems, err := h.itemsRepo.GetByOrderID(int(order.ID))
						if err != nil {
							logrus.Error(fmt.Sprintf("couldn't retrieve existing items for partial order update: %s", err.Error()))
							response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "failed to update order: internal server error")
						} else { // existing items were retrieved successfully (even if there are none)
							// Update existing items for this order, create new items as needed
							for _, item := range order.Items {
								match := false
								for _, existingItem := range existingItems {
									if item.ProductID == existingItem.ProductID {
										match = true
										item.ID = existingItem.ID
										logrus.Info(fmt.Sprintf("Updating item (id: %d) to %+v", item.ID, item))
										_, err := h.itemsRepo.Update(&item, []string{})
										if err != nil {
											logrus.Error(fmt.Sprintf("error updating item: %s", err.Error()))
											response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "failed to update order: internal server error")
										}
										break // assumption: only one item exists with the given product id
									}
								}
								if !match { // The order has no existing item with that product id
									logrus.Info(fmt.Sprintf("Creating new item: %+v", item))
									item.OrderID = int(order.ID)
									_, err := h.itemsRepo.Create(&item)
									if err != nil {
										logrus.Error(fmt.Sprintf("error creating item: %s", err.Error()))
										response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "failed to update order: internal server error")
									}
								}
							}
						}
					}
				}
			}
		} else { // Full update of the order was requested
			isUpdateValid, err := models.ValidateFruitOrder(&order)
			if !isUpdateValid {
				response = utils.NewJsonResponseWithError(http.StatusBadRequest, err.Error())
			} else { // Full update is valid
				// Validate the full update's items
				itemListIsValid, err := h.validateItems(order.Items)
				if !itemListIsValid {
					msg := "Failed to validate updated order item list: " + err.Error()
					logrus.Error(msg)
					response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: msg}}
				} else { // Items are valid
					// Update the order
					logrus.Info(fmt.Sprintf("Trying to update order (id: %d) to %+v", order.ID, order))
					_, err := h.repo.Update(&order, []string{})
					if err != nil {
						logrus.Error(fmt.Sprintf("Error updating order (id: %d) to %+v: %s", order.ID, order, err.Error()))
						response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update order."}}
					} else { // Update was successful
						logrus.Info(fmt.Sprintf("Successfully updated order (id: %d) to %+v", order.ID, order))
						response = utils.JsonResponse{Data: []*models.Order{&order}, Id: strconv.Itoa(int(order.ID))}

						// Get any already existing items for this order
						existingItems, err := h.itemsRepo.GetByOrderID(int(order.ID))
						if err != nil {
							logrus.Error(fmt.Sprintf("couldn't retrieve existing items for a full order update: %s", err.Error()))
							response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "failed to update order: internal server error")
						} else { // existing items were retrieved successfully (even if there are none)
							// Delete any existing items for this order
							for _, item := range existingItems {
								logrus.Info(fmt.Sprintf("deleting existing item (id: %d) from order (id: %d)", item.ID, order.ID))
								_, err := h.itemsRepo.Delete(int64(item.ID))
								if err != nil {
									logrus.Error(fmt.Sprintf("couldn't delete an existing item for a full order update: %s", err.Error()))
									response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "failed to update order: internal server error")
									break
								}
							}
							// Create new items if any exist in this order
							for _, item := range order.Items {
								logrus.Info(fmt.Sprintf("creating new item for order (id: %d)", order.ID))
								_, err := h.itemsRepo.Create(&item)
								if err != nil {
									logrus.Error(fmt.Sprintf("couldn't create an item for a full order update: %s", err.Error()))
									response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "failed to update order: internal server error")
									break
								}
							}
							logrus.Info(fmt.Sprintf("Successfully updated items for order (id: %d)", order.ID))
						}
					}
				}
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else { // Update was successful
		utils.WriteJSONResponse(w, http.StatusOK, response)
	}
}

const (
	idParamName = "id"
)

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

	// Validate order id query parameter
	id, err := utils.GetQueryParamAsInt(r, idParamName)
	if err != nil {
		response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}}
	}

	if response.Error == nil { // Order id is valid
		// Get slice of already existing items for this order
		existingItems, err := h.itemsRepo.GetByOrderID(id)
		if err != nil {
			logrus.Error(fmt.Sprintf("delete order: couldn't retrieve existing items for order (id: %d): %s", id, err.Error()))
			response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "failed to delete order: internal server error")
		} else { // existing items were retrieved successfully (even if there are none)
			// Delete any existing items for this order
			for _, item := range existingItems {
				logrus.Info(fmt.Sprintf("deleting existing item (id: %d) from order (id: %d)", item.ID, id))
				_, err := h.itemsRepo.Delete(int64(item.ID))
				if err != nil {
					logrus.Error(fmt.Sprintf("delete order: couldn't delete an existing item for order (id: %d): %s", id, err.Error()))
					response = utils.NewJsonResponseWithError(http.StatusInternalServerError, "failed to delete order: internal server error")
					break
				}
			}
			logrus.Info(fmt.Sprintf("Successfully deleted items for order (id: %d)", id))

			// Delete the order
			logrus.Info("Deleting order with id " + strconv.Itoa(id) + "...")
			_, err := h.repo.Delete(int64(id))
			if err != nil {
				logrus.Error(fmt.Sprintf("Error deleting order with id = %d: %s", id, err.Error()))
				response = utils.JsonResponse{Error: &utils.JsonErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete order with id = " + strconv.Itoa(id)}}
			} else { // delete was successful
				logrus.Info(fmt.Sprintf("Successfully deleted order with id = %d.", id))
			}
		}
	}

	if response.Error != nil {
		utils.WriteJSONResponse(w, response.Error.Code, response)
	} else { // Delete was successful
		utils.WriteJSONResponse(w, http.StatusNoContent, response) // TODO: This should return no content, need to add another utility function
	}
}

func (h *Order) validateItemsProductID(items []models.Item) (bool, error) {
	idMap := make(map[int]bool, len(items))
	for _, item := range items {
		exists, err := h.productsRepo.Exists(item.ProductID)
		if err != nil {
			return false, errors.New("failed to validate product id: " + err.Error())
		}
		if !exists {
			return false, errors.New("product ID " + strconv.Itoa(item.ProductID) + " does not exist in the repo")
		}
		_, idAlreadyExists := idMap[item.ProductID]
		if idAlreadyExists {
			return false, errors.New("item list contains duplicate product ID: " + strconv.Itoa(int(item.ProductID)))
		}
		idMap[item.ProductID] = true
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
		product, err := h.productsRepo.GetByID(int(item.ProductID))
		if err != nil {
			return -1, err
		}
		subtotal += float64(item.Quantity) * product.Price
	}
	return subtotal, nil
}
