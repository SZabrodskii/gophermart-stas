package services

import (
	"testing"

	"github.com/SZabrodskii/gophermart-stas/internal/utils"
)

func TestValidateOrderNumber(t *testing.T) {
	tests := []struct {
		name        string
		orderNumber string
		valid       bool
	}{
		{"valid luhn", "4561261212345467", true},
		{"invalid luhn", "4561261212345468", false},
		{"too short", "123", false},
		{"non-numeric", "abcd", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ValidateOrderNumber(tt.orderNumber)

			if tt.valid && !result {
				t.Errorf("Expected valid order number, got invalid")
			}
			if !tt.valid && result {
				t.Errorf("Expected invalid order number, got valid")
			}
		})
	}
}
