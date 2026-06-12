package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ungerr"
)

// NewAuthHooks builds the application-specific auth hooks that wire Cashus
// business logic into the generic auth flows. Hooks that have no current
// implementation are left as nil (no-op).
func NewAuthHooks(
	pushNotification service.PushNotificationService,
	profileService service.ProfileService,
	friendshipService service.FriendshipService,
) service.AuthHooks {
	return service.AuthHooks{
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
			profile, err := profileService.GetByID(ctx, uid)
			if err != nil {
				return nil, err
			}
			baseClaims[appconstant.ContextProfileID.String()] = profile.ID
			return baseClaims, nil
		},
	}
}

// associateBySlug links a newly verified user's profile to an anonymous
// profile identified by a slug. It looks up the anonymous profile, finds its
// owner through the friendship graph, creates a real friendship, and then
// associates the profiles.
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

	// Find the owner of the anonymous profile via friendship
	friendships, err := friendshipSvc.GetAll(ctx, anonProfile.ID)
	if err != nil {
		return err
	}
	if len(friendships) == 0 {
		return ungerr.NotFoundError(fmt.Sprintf("no friendship found for slug %s", slug))
	}

	ownerProfileID := friendships[0].ProfileID

	// Create real friendship between owner and new user
	_, err = friendshipSvc.CreateReal(ctx, ownerProfileID, newProfileID)
	if err != nil {
		return err
	}

	// Create RelatedProfile association
	return profileSvc.Associate(ctx, ownerProfileID, newProfileID, anonProfile.ID)
}
