// Package generator provides a cryptographically secure password generator.
package generator

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	lower   = "abcdefghijklmnopqrstuvwxyz"
	upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits  = "0123456789"
	symbols = "!@#$%^&*()-_=+[]{}|;:,.<>?"
)

// Options controls what character classes are included.
type Options struct {
	Length      int
	UseUpper    bool
	UseDigits   bool
	UseSymbols  bool
}

// DefaultOptions returns sensible defaults for a strong password.
func DefaultOptions() Options {
	return Options{
		Length:     20,
		UseUpper:   true,
		UseDigits:  true,
		UseSymbols: true,
	}
}

// Generate creates a random password according to opts.
// It guarantees at least one character from each enabled class.
func Generate(opts Options) (string, error) {
	if opts.Length < 4 {
		opts.Length = 4
	}

	charset := lower
	if opts.UseUpper {
		charset += upper
	}
	if opts.UseDigits {
		charset += digits
	}
	if opts.UseSymbols {
		charset += symbols
	}

	// Build mandatory characters first.
	mandatory := []byte{}
	mandatory = append(mandatory, mustRandChar(lower))
	if opts.UseUpper {
		mandatory = append(mandatory, mustRandChar(upper))
	}
	if opts.UseDigits {
		mandatory = append(mandatory, mustRandChar(digits))
	}
	if opts.UseSymbols {
		mandatory = append(mandatory, mustRandChar(symbols))
	}

	// Fill the rest randomly.
	rest := make([]byte, opts.Length-len(mandatory))
	for i := range rest {
		c, err := randChar(charset)
		if err != nil {
			return "", err
		}
		rest[i] = c
	}

	// Combine and shuffle.
	combined := append(mandatory, rest...)
	if err := shuffle(combined); err != nil {
		return "", err
	}

	return string(combined), nil
}

// Strength returns a qualitative label for a password.
func Strength(pw string) string {
	score := 0
	if len(pw) >= 12 {
		score++
	}
	if len(pw) >= 20 {
		score++
	}
	if strings.ContainsAny(pw, upper) {
		score++
	}
	if strings.ContainsAny(pw, digits) {
		score++
	}
	if strings.ContainsAny(pw, symbols) {
		score++
	}

	switch {
	case score >= 5:
		return "Strong"
	case score >= 3:
		return "Medium"
	default:
		return "Weak"
	}
}

func randChar(charset string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

func mustRandChar(charset string) byte {
	c, _ := randChar(charset)
	return c
}

func shuffle(b []byte) error {
	for i := len(b) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		b[i], b[j.Int64()] = b[j.Int64()], b[i]
	}
	return nil
}
