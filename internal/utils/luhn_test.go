package utils

import "testing"

func TestValidateOrderNumber(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "Valid Luhn number from spec example",
			number:   "12345678903",
			expected: true,
		},
		{
			name:     "Another valid Luhn number",
			number:   "4532015112830366",
			expected: true,
		},
		{
			name:     "Invalid Luhn number",
			number:   "12345678902",
			expected: false,
		},
		{
			name:     "Number with spaces (should be cleaned)",
			number:   "1234 5678 903",
			expected: true,
		},
		{
			name:     "Empty string",
			number:   "",
			expected: false,
		},
		{
			name:     "Non-numeric string",
			number:   "12345abc",
			expected: false,
		},
		{
			name:     "Single digit",
			number:   "0",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateOrderNumber(tt.number)
			if result != tt.expected {
				t.Errorf("ValidateOrderNumber(%q) = %v, want %v", tt.number, result, tt.expected)
			}
		})
	}
}
