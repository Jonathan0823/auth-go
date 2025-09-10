package utils

import (
	"errors"

	"github.com/lib/pq"
)

func IsPGUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// lib/pq
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == "23505"
	}
	return false
}
