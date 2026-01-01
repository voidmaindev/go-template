package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// CentsToString converts cents to a decimal string (e.g., 1999 -> "19.99")
func CentsToString(cents int64) string {
	if cents < 0 {
		return "-" + CentsToString(-cents)
	}
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

// StringToCents converts a decimal string to cents (e.g., "19.99" -> 1999)
func StringToCents(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	// Remove currency symbols if present
	s = strings.TrimLeft(s, "$")
	s = strings.TrimSpace(s)

	// Parse the decimal value
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid money format: %s", s)
	}

	// Convert to cents with proper rounding
	cents := int64(math.Round(f * 100))

	if negative {
		cents = -cents
	}

	return cents, nil
}

// CentsToFloat converts cents to a float64 (e.g., 1999 -> 19.99)
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100
}

// FloatToCents converts a float64 to cents (e.g., 19.99 -> 1999)
func FloatToCents(f float64) int64 {
	return int64(math.Round(f * 100))
}

// FormatMoney formats cents as a money string with currency symbol
func FormatMoney(cents int64, symbol string) string {
	if symbol == "" {
		symbol = "$"
	}
	if cents < 0 {
		return "-" + symbol + CentsToString(-cents)
	}
	return symbol + CentsToString(cents)
}

// AddCents safely adds two cent values
func AddCents(a, b int64) int64 {
	return a + b
}

// SubtractCents safely subtracts two cent values
func SubtractCents(a, b int64) int64 {
	return a - b
}

// MultiplyCents multiplies cents by a quantity
func MultiplyCents(cents int64, quantity int) int64 {
	return cents * int64(quantity)
}

// CalculateTotal calculates total from price and quantity
func CalculateTotal(priceInCents int64, quantity int) int64 {
	return priceInCents * int64(quantity)
}

// CalculatePercentage calculates a percentage of a cent value
func CalculatePercentage(cents int64, percentage float64) int64 {
	return int64(math.Round(float64(cents) * percentage / 100))
}

// ApplyDiscount applies a percentage discount to a cent value
func ApplyDiscount(cents int64, discountPercent float64) int64 {
	discount := CalculatePercentage(cents, discountPercent)
	return cents - discount
}

// ApplyTax applies a tax percentage to a cent value
func ApplyTax(cents int64, taxPercent float64) int64 {
	tax := CalculatePercentage(cents, taxPercent)
	return cents + tax
}
