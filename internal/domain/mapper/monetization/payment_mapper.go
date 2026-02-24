package monetization

import (
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
)

func PaymentToResponse(p entity.Payment) dto.PaymentResponse {
	return dto.PaymentResponse{
		BaseDTO:               mapper.BaseToDTO(p.BaseEntity),
		SubscriptionID:        p.SubscriptionID,
		Amount:                p.Amount,
		Currency:              p.Currency,
		Gateway:               p.Gateway,
		GatewayTransactionID:  p.GatewayTransactionID.String,
		GatewaySubscriptionID: p.GatewaySubscriptionID.String,
		Status:                p.Status,
		FailureReason:         p.FailureReason.String,
		StartsAt:              p.StartsAt.Time,
		EndsAt:                p.EndsAt.Time,
		GatewayEventID:        p.GatewayEventID.String,
		PaidAt:                p.PaidAt.Time,
	}
}
