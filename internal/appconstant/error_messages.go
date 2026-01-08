package appconstant

const (
	ErrAmountMismatched  = "amount mismatch, please check the total amount and the items/fees provided"
	ErrAmountZero        = "amount must be greater than zero"
	ErrNonPositiveAmount = "amount must be positive (>0)"

	ErrNotFriends = "you are not friends with this user, please add them as a friend first"

	ErrProcessFile = "error processing file upload"

	ErrServiceClient = "service client communication failure"

	ErrStructValidation = "error validating struct input"

	ErrAuthUnknownCredentials = "unknown credentials, please check your email/password"

	ErrTransferMethodNotFound = "transfer method with ID: %s is not found"

	ErrDataSelect = "error retrieving data"
	ErrDataUpdate = "error updating data"
	ErrDataInsert = "error inserting new data"
)
