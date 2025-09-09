package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/google/uuid"
)

const CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const SHORTCODE_LENGTH = 8

func GenerateShortCode() (string, error) {
	result := make([]byte, SHORTCODE_LENGTH)
	for i := 0; i < SHORTCODE_LENGTH; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(CHARSET))))
		if err != nil {
			return "", err
		}

		result[i] = CHARSET[num.Int64()]
	}

	return string(result), nil
}

func ParseUUID(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID: %s", id)
	}

	return uid, nil
}
