package friendship

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cocoon-protos/gen/go/friendship/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/domain/profile"
	"github.com/rotisserie/eris"
)

func fromRequestProto(r *friendship.Request) (Request, error) {
	if r == nil {
		return Request{}, eris.New("request is nil")
	}

	id, err := ezutil.Parse[uuid.UUID](r.GetId())
	if err != nil {
		return Request{}, err
	}

	sender, err := profile.FromProfileProto(r.GetSender())
	if err != nil {
		return Request{}, err
	}

	recipient, err := profile.FromProfileProto(r.GetRecipient())
	if err != nil {
		return Request{}, err
	}

	return Request{
		ID:        id,
		Sender:    sender,
		Recipient: recipient,
		Message:   r.GetMessage(),
		CreatedAt: ezutil.FromProtoTime(r.GetCreatedAt()),
		BlockedAt: ezutil.FromProtoTime(r.GetBlockedAt()),
	}, nil
}
