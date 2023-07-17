package util

import (
	"context"
	"fmt"

	"sonamusica-backend/errs"
)

type HasInt64ID interface {
	GetInt64ID() int64
}

// ValidateUpdateSpecs compares len(specs) to the output of dbCounterFunc, and returns error if they differ.
//
// This method is useful for checking the existence of records (from specs) in the DB.
// Commonly, DB update queries use (UPDATE .. SET .. WHERE ID = 1) and doesn't return error if the ID doesn't exist.
// We validate the requested IDs existence using this method.
func ValidateUpdateSpecs[T HasInt64ID](ctx context.Context, specs []T, dbCounterFunc func(context.Context, []int64) (int64, error)) error {
	idsInt := make([]int64, 0, len(specs))
	for _, spec := range specs {
		idsInt = append(idsInt, spec.GetInt64ID())
	}
	count, err := dbCounterFunc(ctx, idsInt)
	if err != nil {
		return fmt.Errorf("dbCounterFunc(): %w", err)
	}
	if count != int64(len(specs)) {
		return errs.NewValidationError(fmt.Errorf("specs length != count of existing IDs in the database (%d vs %d)", len(specs), count),
			errs.ValidationErrorDetail{errs.ClientMessageKey_NonField: "one or more specs have invalid ID"})
	}

	return nil
}
