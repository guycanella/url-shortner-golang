package utils

import "testing"

func TestGenerateShortCode(t *testing.T) {
	code, err := GenerateShortCode()
	if err != nil {
		t.Fatalf("Error to generate short code: %v", err)
	}

	t.Logf("Generated shortcode: %s", code)
}
