package utils

import (
	"testing"
)

func TestCentsToString(t *testing.T) {
	tests := []struct {
		cents    int64
		expected string
	}{
		{0, "0.00"},
		{1, "0.01"},
		{10, "0.10"},
		{99, "0.99"},
		{100, "1.00"},
		{199, "1.99"},
		{1999, "19.99"},
		{10000, "100.00"},
		{123456, "1234.56"},
		{-100, "-1.00"},
		{-199, "-1.99"},
		{-1, "-0.01"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := CentsToString(tt.cents)
			if result != tt.expected {
				t.Errorf("CentsToString(%d) = %q, want %q", tt.cents, result, tt.expected)
			}
		})
	}
}

func TestStringToCents(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"0", 0, false},
		{"0.00", 0, false},
		{"0.01", 1, false},
		{"0.10", 10, false},
		{"0.99", 99, false},
		{"1.00", 100, false},
		{"1.99", 199, false},
		{"19.99", 1999, false},
		{"100.00", 10000, false},
		{"1234.56", 123456, false},
		{"-1.00", -100, false},
		{"-19.99", -1999, false},
		{"$19.99", 1999, false},
		{" 19.99 ", 1999, false},
		{"", 0, false},
		{"   ", 0, false},
		// Rounding
		{"1.999", 200, false},
		{"1.994", 199, false},
		{"1.995", 200, false}, // rounds up
		// Invalid
		{"abc", 0, true},
		{"1.2.3", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := StringToCents(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("StringToCents(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("StringToCents(%q) unexpected error = %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("StringToCents(%q) = %d, want %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestCentsToFloat(t *testing.T) {
	tests := []struct {
		cents    int64
		expected float64
	}{
		{0, 0.0},
		{100, 1.0},
		{199, 1.99},
		{1999, 19.99},
		{-100, -1.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := CentsToFloat(tt.cents)
			if result != tt.expected {
				t.Errorf("CentsToFloat(%d) = %f, want %f", tt.cents, result, tt.expected)
			}
		})
	}
}

func TestFloatToCents(t *testing.T) {
	tests := []struct {
		input    float64
		expected int64
	}{
		{0.0, 0},
		{1.0, 100},
		{1.99, 199},
		{19.99, 1999},
		{-1.0, -100},
		{1.994, 199}, // rounds down
		{1.995, 200}, // rounds up (banker's rounding)
		{1.999, 200},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := FloatToCents(tt.input)
			if result != tt.expected {
				t.Errorf("FloatToCents(%f) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		cents    int64
		symbol   string
		expected string
	}{
		{1999, "$", "$19.99"},
		{1999, "", "$19.99"},    // default symbol
		{1999, "€", "€19.99"},
		{1999, "£", "£19.99"},
		{0, "$", "$0.00"},
		{-1999, "$", "-$19.99"},
		{-1, "$", "-$0.01"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatMoney(tt.cents, tt.symbol)
			if result != tt.expected {
				t.Errorf("FormatMoney(%d, %q) = %q, want %q", tt.cents, tt.symbol, result, tt.expected)
			}
		})
	}
}

func TestAddCents(t *testing.T) {
	tests := []struct {
		a, b     int64
		expected int64
	}{
		{100, 200, 300},
		{0, 100, 100},
		{100, 0, 100},
		{-100, 200, 100},
		{100, -200, -100},
	}

	for _, tt := range tests {
		result := AddCents(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("AddCents(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestSubtractCents(t *testing.T) {
	tests := []struct {
		a, b     int64
		expected int64
	}{
		{300, 200, 100},
		{100, 100, 0},
		{100, 200, -100},
		{-100, -200, 100},
	}

	for _, tt := range tests {
		result := SubtractCents(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("SubtractCents(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMultiplyCents(t *testing.T) {
	tests := []struct {
		cents    int64
		quantity int
		expected int64
	}{
		{100, 3, 300},
		{199, 2, 398},
		{100, 0, 0},
		{100, -1, -100},
	}

	for _, tt := range tests {
		result := MultiplyCents(tt.cents, tt.quantity)
		if result != tt.expected {
			t.Errorf("MultiplyCents(%d, %d) = %d, want %d", tt.cents, tt.quantity, result, tt.expected)
		}
	}
}

func TestCalculateTotal(t *testing.T) {
	tests := []struct {
		price    int64
		quantity int
		expected int64
	}{
		{1999, 1, 1999},
		{1999, 2, 3998},
		{1999, 10, 19990},
		{0, 10, 0},
		{1999, 0, 0},
	}

	for _, tt := range tests {
		result := CalculateTotal(tt.price, tt.quantity)
		if result != tt.expected {
			t.Errorf("CalculateTotal(%d, %d) = %d, want %d", tt.price, tt.quantity, result, tt.expected)
		}
	}
}

func TestCalculatePercentage(t *testing.T) {
	tests := []struct {
		cents      int64
		percentage float64
		expected   int64
	}{
		{10000, 10, 1000},   // 10% of $100 = $10
		{10000, 50, 5000},   // 50% of $100 = $50
		{10000, 100, 10000}, // 100% of $100 = $100
		{10000, 0, 0},       // 0% of $100 = $0
		{1999, 10, 200},     // 10% of $19.99 = $2.00 (rounded)
		{1999, 5, 100},      // 5% of $19.99 = $1.00 (rounded)
	}

	for _, tt := range tests {
		result := CalculatePercentage(tt.cents, tt.percentage)
		if result != tt.expected {
			t.Errorf("CalculatePercentage(%d, %f) = %d, want %d", tt.cents, tt.percentage, result, tt.expected)
		}
	}
}

func TestApplyDiscount(t *testing.T) {
	tests := []struct {
		cents    int64
		discount float64
		expected int64
	}{
		{10000, 10, 9000},  // 10% off $100 = $90
		{10000, 20, 8000},  // 20% off $100 = $80
		{10000, 50, 5000},  // 50% off $100 = $50
		{10000, 100, 0},    // 100% off $100 = $0
		{10000, 0, 10000},  // 0% off $100 = $100
		{1999, 10, 1799},   // 10% off $19.99 = $17.99
	}

	for _, tt := range tests {
		result := ApplyDiscount(tt.cents, tt.discount)
		if result != tt.expected {
			t.Errorf("ApplyDiscount(%d, %f) = %d, want %d", tt.cents, tt.discount, result, tt.expected)
		}
	}
}

func TestApplyTax(t *testing.T) {
	tests := []struct {
		cents    int64
		tax      float64
		expected int64
	}{
		{10000, 10, 11000},  // $100 + 10% tax = $110
		{10000, 20, 12000},  // $100 + 20% tax = $120
		{10000, 0, 10000},   // $100 + 0% tax = $100
		{1999, 10, 2199},    // $19.99 + 10% tax = $21.99
	}

	for _, tt := range tests {
		result := ApplyTax(tt.cents, tt.tax)
		if result != tt.expected {
			t.Errorf("ApplyTax(%d, %f) = %d, want %d", tt.cents, tt.tax, result, tt.expected)
		}
	}
}

func TestRoundTrip_CentsToStringToBack(t *testing.T) {
	// Test that converting to string and back preserves value
	testCases := []int64{0, 1, 99, 100, 199, 1999, 123456, -100, -1999}

	for _, cents := range testCases {
		str := CentsToString(cents)
		result, err := StringToCents(str)
		if err != nil {
			t.Errorf("Round trip failed for %d: %v", cents, err)
		}
		if result != cents {
			t.Errorf("Round trip: %d -> %q -> %d", cents, str, result)
		}
	}
}
