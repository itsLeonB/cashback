package dto

type TransferMethodResponse struct {
	BaseDTO
	Name    string `json:"name"`
	Display string `json:"display"`
}
