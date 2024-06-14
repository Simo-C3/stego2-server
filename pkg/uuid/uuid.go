package uuid

import (
	"github.com/google/uuid"
)

func GenerateUUIDv7() (string, error) {
	uuid7, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	return uuid7.String(), nil
}
