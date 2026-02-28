package encoder

import (
	"errors"
	"strings"
)

// Base62Encoder handles encoding and decoding of integers to Base62 strings
type Base62Encoder struct {
	base62Chars string
	base        int
}

// NewBase62Encoder creates a new Base62 encoder instance
func NewBase62Encoder() *Base62Encoder {
	return &Base62Encoder{
		base62Chars: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		base:        62,
	}
}

// Encode converts an integer to a Base62 string
// Example: 125 -> "2D"
// Time Complexity: O(log62(num))
// Space Complexity: O(log62(num))
func (e *Base62Encoder) Encode(num int64) string {
	if num == 0 {
		return string(e.base62Chars[0])
	}

	var result strings.Builder
	for num > 0 {
		remainder := num % int64(e.base)
		result.WriteByte(e.base62Chars[remainder])
		num /= int64(e.base)
	}

	// Reverse the string
	encoded := result.String()
	return reverse(encoded)
}

// Decode converts a Base62 string back to an integer
// Example: "2D" -> 125
// Time Complexity: O(len(shortCode))
// Space Complexity: O(1)
func (e *Base62Encoder) Decode(shortCode string) (int64, error) {
	if shortCode == "" {
		return 0, errors.New("cannot decode empty string")
	}

	var num int64
	for _, char := range shortCode {
		index := strings.IndexRune(e.base62Chars, char)
		if index == -1 {
			return 0, errors.New("invalid character in short code")
		}
		num = num*int64(e.base) + int64(index)
	}

	return num, nil
}

// GenerateShortCode generates a short code with minimum length
func (e *Base62Encoder) GenerateShortCode(num int64, minLength int) string {
	encoded := e.Encode(num)
	if len(encoded) < minLength {
		// Pad with zeros
		padding := strings.Repeat("0", minLength-len(encoded))
		return padding + encoded
	}
	return encoded
}

// Helper function to reverse a string
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
