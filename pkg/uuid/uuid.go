package uuid

import (
	"github.com/google/uuid"
)

func GenerateUUIDv7() string {
	uuid7, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	return uuid7.String()
}
