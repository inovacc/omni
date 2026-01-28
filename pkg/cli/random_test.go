package cli

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestRunRandom(t *testing.T) {
	t.Run("default string", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRandom(&buf, RandomOptions{})
		if err != nil {
			t.Fatalf("RunRandom() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if len(result) != 16 {
			t.Errorf("RunRandom() length = %d, want 16", len(result))
		}
	})

	t.Run("custom length", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRandom(&buf, RandomOptions{Length: 32})
		if err != nil {
			t.Fatalf("RunRandom() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if len(result) != 32 {
			t.Errorf("RunRandom() length = %d, want 32", len(result))
		}
	})

	t.Run("multiple values", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRandom(&buf, RandomOptions{Count: 5, Length: 8})
		if err != nil {
			t.Fatalf("RunRandom() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunRandom() count = %d, want 5", len(lines))
		}
	})

	t.Run("integer type", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRandom(&buf, RandomOptions{Type: "int", Min: 0, Max: 100})
		if err != nil {
			t.Fatalf("RunRandom() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())

		num, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			t.Fatalf("RunRandom() not a number: %v", result)
		}

		if num < 0 || num >= 100 {
			t.Errorf("RunRandom() = %d, want 0-99", num)
		}
	})

	t.Run("hex type", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRandom(&buf, RandomOptions{Type: "hex", Length: 16})
		if err != nil {
			t.Fatalf("RunRandom() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())

		matched, _ := regexp.MatchString("^[0-9a-f]+$", result)
		if !matched {
			t.Errorf("RunRandom() = %v, not valid hex", result)
		}
	})

	t.Run("alpha type", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRandom(&buf, RandomOptions{Type: "alpha", Length: 20})
		if err != nil {
			t.Fatalf("RunRandom() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())

		matched, _ := regexp.MatchString("^[a-zA-Z]+$", result)
		if !matched {
			t.Errorf("RunRandom() = %v, not valid alpha", result)
		}
	})

	t.Run("password type", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRandom(&buf, RandomOptions{Type: "password", Length: 20})
		if err != nil {
			t.Fatalf("RunRandom() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if len(result) != 20 {
			t.Errorf("RunRandom() length = %d, want 20", len(result))
		}
	})

	t.Run("float type", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRandom(&buf, RandomOptions{Type: "float"})
		if err != nil {
			t.Fatalf("RunRandom() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())

		f, err := strconv.ParseFloat(result, 64)
		if err != nil {
			t.Fatalf("RunRandom() not a float: %v", result)
		}

		if f < 0 || f > 1 {
			t.Errorf("RunRandom() = %f, want 0-1", f)
		}
	})

	t.Run("unknown type", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRandom(&buf, RandomOptions{Type: "unknown"})
		if err == nil {
			t.Error("RunRandom() expected error for unknown type")
		}
	})
}

func TestRandomString(t *testing.T) {
	result := RandomString(20)
	if len(result) != 20 {
		t.Errorf("RandomString() length = %d, want 20", len(result))
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", result)
	if !matched {
		t.Errorf("RandomString() = %v, not alphanumeric", result)
	}
}

func TestRandomHex(t *testing.T) {
	result := RandomHex(16)
	if len(result) != 16 {
		t.Errorf("RandomHex() length = %d, want 16", len(result))
	}

	matched, _ := regexp.MatchString("^[0-9a-f]+$", result)
	if !matched {
		t.Errorf("RandomHex() = %v, not hex", result)
	}
}

func TestRandomPassword(t *testing.T) {
	result := RandomPassword(20)
	if len(result) != 20 {
		t.Errorf("RandomPassword() length = %d, want 20", len(result))
	}
}

func TestRandomInt(t *testing.T) {
	for range 100 {
		result := RandomInt(10, 20)
		if result < 10 || result >= 20 {
			t.Errorf("RandomInt(10, 20) = %d, out of range", result)
		}
	}

	// Test edge case where max <= min
	result := RandomInt(10, 5)
	if result != 10 {
		t.Errorf("RandomInt(10, 5) = %d, want 10", result)
	}
}

func TestRandomChoice(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	counts := make(map[string]int)

	for range 1000 {
		choice := RandomChoice(items)
		counts[choice]++
	}

	// All items should be chosen at least once
	for _, item := range items {
		if counts[item] == 0 {
			t.Errorf("RandomChoice() never chose %v", item)
		}
	}

	// Empty slice should return zero value
	var empty []string

	result := RandomChoice(empty)
	if result != "" {
		t.Errorf("RandomChoice([]) = %v, want empty string", result)
	}
}

func TestShuffle(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	original := make([]int, len(items))
	copy(original, items)

	Shuffle(items)

	// Should contain same elements
	sum := 0
	for _, v := range items {
		sum += v
	}

	if sum != 55 {
		t.Error("Shuffle() lost elements")
	}

	// Should be different order (with very high probability)
	same := true

	for i := range items {
		if items[i] != original[i] {
			same = false
			break
		}
	}

	if same {
		t.Log("Shuffle() returned same order (unlikely but possible)")
	}
}
