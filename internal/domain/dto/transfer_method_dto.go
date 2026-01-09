package dto

import "github.com/google/uuid"

type TransferMethodResponse struct {
	BaseDTO
	Name     string    `json:"name"`
	Display  string    `json:"display"`
	IconURL  string    `json:"iconUrl"`
	ParentID uuid.UUID `json:"parentId"`
}
