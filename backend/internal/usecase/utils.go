package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// toText converts *string to pgtype.Text
func toText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// toTextFromStr converts string to pgtype.Text (empty string check optional)
func toTextFromStr(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// toTimestamp converts *time.Time to pgtype.Timestamptz
func toTimestamp(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// toUUID converts *uuid.UUID to pgtype.UUID
func toUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}
