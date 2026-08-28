package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/stretchr/testify/assert"
)

func realProfile() profileLookup {
	return profileLookup{found: true, profile: users.UserProfile{UserID: uuid.NullUUID{UUID: uuid.New(), Valid: true}}}
}

func anonProfile() profileLookup {
	return profileLookup{found: true, profile: users.UserProfile{}}
}

func TestDecideRow(t *testing.T) {
	realID := uuid.New()
	anonID := uuid.New()
	row := relatedProfileRow{RealProfileID: realID, AnonProfileID: anonID}

	tests := []struct {
		name       string
		row        relatedProfileRow
		anon, real profileLookup
		wantAction rowAction
	}{
		{
			name:       "garbage row: real == anon",
			row:        relatedProfileRow{RealProfileID: anonID, AnonProfileID: anonID},
			anon:       anonProfile(),
			real:       anonProfile(),
			wantAction: actionFlagManualReview,
		},
		{
			name:       "anon profile already gone",
			row:        row,
			anon:       profileLookup{found: false},
			real:       realProfile(),
			wantAction: actionSkipAlreadyGone,
		},
		{
			name:       "anon profile already real (data corruption)",
			row:        row,
			anon:       realProfile(),
			real:       realProfile(),
			wantAction: actionFlagManualReview,
		},
		{
			name:       "real profile missing",
			row:        row,
			anon:       anonProfile(),
			real:       profileLookup{found: false},
			wantAction: actionFlagManualReview,
		},
		{
			name:       "real profile not actually real",
			row:        row,
			anon:       anonProfile(),
			real:       anonProfile(),
			wantAction: actionFlagManualReview,
		},
		{
			name:       "normal case: merge",
			row:        row,
			anon:       anonProfile(),
			real:       realProfile(),
			wantAction: actionMerge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decideRow(tt.row, tt.anon, tt.real)
			assert.Equal(t, tt.wantAction, got.action)
		})
	}
}

func TestLookupProfile_NilUUID(t *testing.T) {
	// A nil svcs is safe here only because the uuid.Nil check must
	// short-circuit before ever touching svcs.
	got, err := lookupProfile(context.Background(), nil, uuid.Nil)

	assert.NoError(t, err)
	assert.False(t, got.found)
}
