package appconstant

const (
	ErrAmountMismatched = "amount mismatch, please check the total amount and the items/fees provided"
	ErrAmountZero       = "amount must be greater than zero"

	ErrNotFriends = "you are not friends with this user, please add them as a friend first"

	ErrTransferMethodNotFound = "transfer method with ID: %s is not found"

	ErrDataSelect = "error retrieving data"
	ErrDataUpdate = "error updating data"
	ErrDataInsert = "error inserting new data"
)
