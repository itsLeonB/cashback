// Command backfill-related-profiles is a one-off operator tool for repointing
// pre-existing related_profiles rows left over from before the anon-profile
// physical-merge refactor (see docs/anon-profile-merge-verification.md).
//
// For each related_profiles row it repoints every FK reference from
// anon_profile_id to real_profile_id and deletes the anon profile, using the
// same repository-level operations as ProfileService.MergeAnonymousProfile
// (internal/domain/service/profile_service.go) minus its friendship
// authorization gate, which only makes sense for a live, user-initiated HTTP
// request, not an operator backfill.
//
// Usage:
//
//	go run ./cmd/backfill-related-profiles            # dry-run, no writes
//	go run ./cmd/backfill-related-profiles --apply     # perform the repoints
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/ungerr"
	_ "github.com/joho/godotenv/autoload"
)

type relatedProfileRow struct {
	RealProfileID uuid.UUID `gorm:"column:real_profile_id"`
	AnonProfileID uuid.UUID `gorm:"column:anon_profile_id"`
}

type rowAction int

const (
	actionMerge rowAction = iota
	actionSkipAlreadyGone
	actionFlagManualReview
)

type rowResult struct {
	action rowAction
	reason string
}

type profileLookup struct {
	profile users.UserProfile
	found   bool
}

func main() {
	apply := flag.Bool("apply", false, "actually perform repoints; without this, dry-run only")
	flag.Parse()

	logger.Init("BackfillRelatedProfiles")

	if err := config.Load(); err != nil {
		logger.Fatal(err)
	}

	providers, err := provider.All()
	if err != nil {
		logger.Fatal(err)
	}
	defer func() {
		if err := providers.Shutdown(); err != nil {
			logger.Errorf("error shutting down providers: %v", err)
		}
	}()

	ctx := context.Background()
	repos := providers.Repositories
	svcs := providers.Services

	var rows []relatedProfileRow
	if err := providers.DataSources.Gorm.WithContext(ctx).
		Raw("SELECT real_profile_id, anon_profile_id FROM related_profiles").
		Scan(&rows).Error; err != nil {
		logger.Fatal(err)
	}
	logger.Infof("found %d related_profiles row(s) to process", len(rows))

	// Row order doesn't matter: a profile's IsReal() (its UserID linkage) is
	// never mutated by any Repoint* call, and related_profiles creation
	// already guarantees anon_profile_id can't be another row's
	// real_profile_id - so no row's classification can ever depend on
	// another row having run first.
	rpt := &report{}
	for _, row := range rows {
		processRow(ctx, repos, svcs, row, *apply, rpt)
	}

	var remaining int64
	countErr := providers.DataSources.Gorm.WithContext(ctx).
		Raw("SELECT count(*) FROM related_profiles").Scan(&remaining).Error

	if countErr != nil {
		logger.Errorf("failed to verify remaining related_profiles count: %v", countErr)
		logger.Infof("done: merged=%d skipped-already-gone=%d flagged-manual-review=%d failed=%d related_profiles-remaining=unknown (verification query failed)",
			rpt.merged, rpt.skipped, rpt.flagged, rpt.failed)
	} else {
		logger.Infof("done: merged=%d skipped-already-gone=%d flagged-manual-review=%d failed=%d related_profiles-remaining=%d",
			rpt.merged, rpt.skipped, rpt.flagged, rpt.failed, remaining)
	}
	for _, reason := range rpt.flaggedReasons {
		logger.Warnf("  %s", reason)
	}
}

// report tallies outcomes across all rows and centralizes the
// "manual review needed" logging so it isn't repeated at every call site.
type report struct {
	merged, skipped, flagged, failed int
	flaggedReasons                   []string
}

func (r *report) flagManualReview(row relatedProfileRow, reason string) {
	r.flagged++
	msg := fmt.Sprintf("anon=%s real=%s: %s", row.AnonProfileID, row.RealProfileID, reason)
	r.flaggedReasons = append(r.flaggedReasons, msg)
	logger.Warnf("manual review needed: %s", msg)
}

// processRow classifies and, if eligible, repoints a single row, recording
// the outcome on rpt. Split out of main solely to keep the top-level driver
// loop flat.
func processRow(ctx context.Context, repos *provider.Repositories, svcs *provider.Services, row relatedProfileRow, apply bool, rpt *report) {
	anonLookup, err := lookupProfile(ctx, svcs, row.AnonProfileID)
	if err != nil {
		rpt.flagManualReview(row, fmt.Sprintf("error looking up anon profile: %v", err))
		return
	}
	realLookup, err := lookupProfile(ctx, svcs, row.RealProfileID)
	if err != nil {
		rpt.flagManualReview(row, fmt.Sprintf("error looking up real profile: %v", err))
		return
	}

	result := decideRow(row, anonLookup, realLookup)
	switch result.action {
	case actionSkipAlreadyGone:
		rpt.skipped++
		logger.Infof("skip anon=%s real=%s: %s", row.AnonProfileID, row.RealProfileID, result.reason)
		return
	case actionFlagManualReview:
		rpt.flagManualReview(row, result.reason)
		return
	}

	if !apply {
		logger.Infof("[dry-run] would repoint anon=%s -> real=%s", row.AnonProfileID, row.RealProfileID)
		rpt.merged++
		return
	}

	if err := repointRow(ctx, repos, anonLookup.profile, row); err != nil {
		rpt.failed++
		logger.Errorf("failed to repoint anon=%s real=%s: %v", row.AnonProfileID, row.RealProfileID, err)
		return
	}
	rpt.merged++
	logger.Infof("merged anon=%s -> real=%s", row.AnonProfileID, row.RealProfileID)
}

// decideRow classifies a row given already-resolved profile lookups. Pure:
// no DB or service calls, so it's directly unit-testable.
func decideRow(row relatedProfileRow, anon, real profileLookup) rowResult {
	if row.RealProfileID == row.AnonProfileID {
		return rowResult{action: actionFlagManualReview, reason: "garbage row: real_profile_id == anon_profile_id"}
	}
	if !anon.found {
		return rowResult{action: actionSkipAlreadyGone, reason: "anon profile already gone (already merged)"}
	}
	if anon.profile.IsReal() {
		return rowResult{action: actionFlagManualReview, reason: "anon_profile_id already belongs to a real profile"}
	}
	if !real.found {
		return rowResult{action: actionFlagManualReview, reason: "real_profile_id does not exist"}
	}
	if !real.profile.IsReal() {
		return rowResult{action: actionFlagManualReview, reason: "real_profile_id does not belong to a real profile"}
	}
	return rowResult{action: actionMerge}
}

func lookupProfile(ctx context.Context, svcs *provider.Services, id uuid.UUID) (profileLookup, error) {
	if id == uuid.Nil {
		// GetEntityByID builds its WHERE clause from a struct (go-crud's
		// WhereBySpec), which GORM populates by skipping zero-value fields -
		// a nil ID would silently match an arbitrary row instead of "not
		// found", so short-circuit before that ever runs.
		return profileLookup{}, nil
	}
	profile, err := svcs.Profile.GetEntityByID(ctx, id)
	if isNotFound(err) {
		return profileLookup{}, nil
	}
	if err != nil {
		return profileLookup{}, err
	}
	return profileLookup{profile: profile, found: true}, nil
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var appErr ungerr.AppError
	return errors.As(err, &appErr) && appErr.HttpStatus() == http.StatusNotFound
}

// repointRow mirrors ProfileService.MergeAnonymousProfile's repoint-and-delete
// sequence (internal/domain/service/profile_service.go:353-414) exactly,
// minus the two checkFriendship calls - those authorize a live user's HTTP
// request and don't apply to an operator backfill of legacy data.
//
// ponytail: this call list is copied rather than shared via a new exported
// helper - this is a one-shot script meant to run once before cutover, not
// long-lived code; if that table list changes before this runs, update both.
func repointRow(ctx context.Context, repos *provider.Repositories, anonProfile users.UserProfile, row relatedProfileRow) error {
	return repos.Transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := repos.Friendship.RepointFriendships(ctx, row.AnonProfileID, row.RealProfileID); err != nil {
			return err
		}
		if err := repos.FriendshipRequest.RepointProfile(ctx, row.AnonProfileID, row.RealProfileID); err != nil {
			return err
		}
		if err := repos.ProfileTransferMethod.RepointProfile(ctx, row.AnonProfileID, row.RealProfileID); err != nil {
			return err
		}
		if err := repos.DebtTransaction.RepointProfile(ctx, row.AnonProfileID, row.RealProfileID); err != nil {
			return err
		}
		if err := repos.GroupExpense.RepointProfile(ctx, row.AnonProfileID, row.RealProfileID); err != nil {
			return err
		}
		if err := repos.ExpenseItem.RepointParticipants(ctx, row.AnonProfileID, row.RealProfileID); err != nil {
			return err
		}
		if err := repos.OtherFee.RepointParticipants(ctx, row.AnonProfileID, row.RealProfileID); err != nil {
			return err
		}
		if err := repos.Notification.RepointProfile(ctx, row.AnonProfileID, row.RealProfileID); err != nil {
			return err
		}
		if err := repos.PushSubscription.RepointProfile(ctx, row.AnonProfileID, row.RealProfileID); err != nil {
			return err
		}
		// Also purges this row's related_profiles entry as a side effect
		// (internal/adapters/repository/profile_repository.go Delete).
		return repos.Profile.Delete(ctx, anonProfile)
	})
}
