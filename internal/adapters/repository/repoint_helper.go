package repository

import (
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

// findOptional runs query (expected to match at most one row) and scans it into dest, returning
// (false, nil) instead of an error when nothing matches. Every profile-merge repoint method
// treats "the real profile doesn't already have a colliding row" as a normal outcome, not an
// error, so this factors out the gorm.ErrRecordNotFound check repeated at each call site.
func findOptional[T any](query *gorm.DB, dest *T) (bool, error) {
	err := query.Take(dest).Error
	if err == nil {
		return true, nil
	}
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return false, ungerr.Wrap(err, appconstant.ErrDataSelect)
}
