package luhn

import "testing"

// Тестируем функцию ValidateLuhn
func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{"Valid number", "12345678903", true},
		{"Invalid number", "1234567890", false},
		{"Valid number with spaces", "123 456 789 03", false},
		{"Empty string", "", true},
		{"Invalid number with non-digit", "12345A78903", false},
		{"Valid number with long length", "9876543210987654", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateLuhn(tt.number)
			if result != tt.expected {
				t.Errorf("For number %s, expected %v, but got %v", tt.number, tt.expected, result)
			}
		})
	}
}