// Package handler provides implementations for handlers that will handle incoming requests to perform operations on repositories of various data types in the fruitbar application.
package handler

const (
	validationFailedErrMsgPrefix = "Validation failed: "
	internalServerErrMsg         = "Internal server error. Please contact your system administrator."
	unauthorizedErrMsg           = "Authorization failed."
	unauthorizedErrMsgPrefix     = "Authorization failed: "
	forbiddenErrMsgPrefix        = "Forbidden: Not enough privileges to "

	forbiddenReadOrderErrMsg   = forbiddenErrMsgPrefix + "read this Order."
	forbiddenUpdateOrderErrMsg = forbiddenErrMsgPrefix + "update this Order."
	forbiddenDeleteOrderErrMsg = forbiddenErrMsgPrefix + "delete this Order."

	forbiddenCreateUserErrMsg = forbiddenErrMsgPrefix + "create Users with the 'employee' or 'admin' roles."
	forbiddenReadUserErrMsg   = forbiddenErrMsgPrefix + "read this User."
	forbiddenUpdateUserErrMsg = forbiddenErrMsgPrefix + "update this User."
	forbiddenDeleteUserErrMsg = forbiddenErrMsgPrefix + "delete this User."

	forbiddenCreateProductErrMsg = forbiddenErrMsgPrefix + "create a Product."
	forbiddenUpdateProductErrMsg = forbiddenErrMsgPrefix + "update a Product."
	forbiddenDeleteProductErrMsg = forbiddenErrMsgPrefix + "delete a Product."

	idParam     = "id"
	fieldsParam = "fields"

	readPageMaxLimit         = 10
	readOrdersMaxRecordLimit = 10
)
