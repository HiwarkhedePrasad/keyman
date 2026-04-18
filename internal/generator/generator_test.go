package generator_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/keyman/keyman/internal/generator"
)

func TestGenerateLength(t *testing.T) {
	for _, length := range []int{8, 12, 20, 32, 64} {
		opts := generator.DefaultOptions()
		opts.Length = length
		pw, err := generator.Generate(opts)
		if err != nil {
			t.Fatalf("Generate() error for length %d: %v", length, err)
		}
		if len(pw) != length {
			t.Fatalf("expected length %d, got %d", length, len(pw))
		}
	}
}

func TestGenerateContainsUpperWhenEnabled(t *testing.T) {
	opts := generator.DefaultOptions()
	opts.UseUpper = true
	opts.Length = 32

	hasUpper := false
	for i := 0; i < 10; i++ {
		pw, _ := generator.Generate(opts)
		for _, r := range pw {
			if unicode.IsUpper(r) {
				hasUpper = true
			}
		}
	}
	if !hasUpper {
		t.Fatal("expected at least one uppercase character")
	}
}

func TestGenerateNoUpperWhenDisabled(t *testing.T) {
	opts := generator.Options{Length: 32, UseUpper: false, UseDigits: true, UseSymbols: false}
	for i := 0; i < 20; i++ {
		pw, _ := generator.Generate(opts)
		for _, r := range pw {
			if unicode.IsUpper(r) {
				t.Fatalf("found uppercase in password when disabled: %q", pw)
			}
		}
	}
}

func TestGenerateContainsDigitWhenEnabled(t *testing.T) {
	opts := generator.DefaultOptions()
	opts.UseDigits = true
	opts.Length = 32

	for i := 0; i < 20; i++ {
		pw, _ := generator.Generate(opts)
		if strings.ContainsAny(pw, "0123456789") {
			return // pass
		}
	}
	t.Fatal("expected at least one digit after 20 attempts")
}

func TestGenerateNoDigitsWhenDisabled(t *testing.T) {
	opts := generator.Options{Length: 32, UseUpper: true, UseDigits: false, UseSymbols: false}
	for i := 0; i < 20; i++ {
		pw, _ := generator.Generate(opts)
		if strings.ContainsAny(pw, "0123456789") {
			t.Fatalf("found digit in password when disabled: %q", pw)
		}
	}
}

func TestGenerateUnique(t *testing.T) {
	opts := generator.DefaultOptions()
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		pw, _ := generator.Generate(opts)
		if seen[pw] {
			t.Fatalf("generated duplicate password: %q", pw)
		}
		seen[pw] = true
	}
}

func TestStrength(t *testing.T) {
	cases := []struct {
		pw   string
		want string
	}{
		{"abc", "Weak"},
		{"abcdef1234", "Medium"},
		{"Abcdefgh1234!@#$", "Strong"},
	}
	for _, c := range cases {
		got := generator.Strength(c.pw)
		if got != c.want {
			t.Errorf("Strength(%q) = %q, want %q", c.pw, got, c.want)
		}
	}
}
