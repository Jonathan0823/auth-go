package postgres

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func pgtypeUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgtypeUUIDFromString(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("parse uuid: %w", err)
	}
	return pgtypeUUID(id), nil
}

func pgtypeTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: !t.IsZero()}
}
