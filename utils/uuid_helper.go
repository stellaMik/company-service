package utils

import "github.com/google/uuid"

// Generate a UUID using "github.com/google/uuid" library. Returns the generated UUID or an error if occurs.
func GenerateUUID() uuid.UUID {
	return uuid.New()

}

func GenerateUUIDFromString(strUUID string) (uuid.UUID, error) {
	parseUUID, err := uuid.Parse(strUUID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parseUUID, nil
}
