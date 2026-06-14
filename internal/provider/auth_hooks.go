package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/go-authkit"
	"github.com/itsLeonB/ungerr"
)

// NewAuthKitHooks builds authkit.Hooks wiring Cashus business logic.
func NewAuthKitHooks(
	pushNotification service.PushNotificationService,
	profileService service.ProfileService,
	friendshipService service.FriendshipService,
) authkit.Hooks {
	return authkit.Hooks{
		BeforeLogout: func(ctx context.Context, sessionID string) error {
			sid, err := uuid.Parse(sessionID)
			if err != nil {
				return err
			}
			return pushNotification.UnsubscribeBySession(ctx, sid)
		},
		AfterEmailVerified: func(ctx context.Context, userID, profileID string, claims map[string]any) error {
			if profileID == "" {
				return nil
			}
			pid, err := uuid.Parse(profileID)
			if err != nil {
				return err
			}
			slug, ok := claims["slug"].(string)
			if !ok || slug == "" {
				return nil
			}
			return associateBySlug(ctx, profileService, friendshipService, pid, slug)
		},
		ClaimsBuilder: func(ctx context.Context, userID string, baseClaims map[string]any) (map[string]any, error) {
			uid, err := uuid.Parse(userID)
			if err != nil {
				return nil, err
			}
			profileID, err := profileService.GetProfileIDByUserID(ctx, uid)
			if err != nil {
				return nil, err
			}
			baseClaims[appconstant.ContextProfileID.String()] = profileID
			return baseClaims, nil
		},
	}
}

func associateBySlug(
	ctx context.Context,
	profileSvc service.ProfileService,
	friendshipSvc service.FriendshipService,
	newProfileID uuid.UUID,
	slug string,
) error {
	anonProfile, err := profileSvc.FindBySlug(ctx, slug)
	if err != nil {
		return err
	}

	friendships, err := friendshipSvc.GetAll(ctx, anonProfile.ID)
	if err != nil {
		return err
	}
	if len(friendships) == 0 {
		return ungerr.NotFoundError(fmt.Sprintf("no friendship found for slug %s", slug))
	}

	ownerProfileID := friendships[0].ProfileID

	_, err = friendshipSvc.CreateReal(ctx, ownerProfileID, newProfileID)
	if err != nil {
		return err
	}

	return profileSvc.Associate(ctx, ownerProfileID, newProfileID, anonProfile.ID)
}
