package yes

import (
	"testing"
)

func TestYes(t *testing.T) {
	t.Run("default output", func(t *testing.T) {
		result := Yes("", 5)

		if len(result) != 5 {
			t.Errorf("Yes() got %d items, want 5", len(result))
		}

		for i, v := range result {
			if v != "y" {
				t.Errorf("Yes()[%d] = %v, want 'y'", i, v)
			}
		}
	})

	t.Run("custom output", func(t *testing.T) {
		result := Yes("hello", 3)

		if len(result) != 3 {
			t.Errorf("Yes() got %d items, want 3", len(result))
		}

		for i, v := range result {
			if v != "hello" {
				t.Errorf("Yes()[%d] = %v, want 'hello'", i, v)
			}
		}
	})

	t.Run("zero count", func(t *testing.T) {
		result := Yes("y", 0)

		if len(result) != 0 {
			t.Errorf("Yes() with 0 count got %d items, want 0", len(result))
		}
	})

	t.Run("single item", func(t *testing.T) {
		result := Yes("test", 1)

		if len(result) != 1 || result[0] != "test" {
			t.Errorf("Yes() single = %v", result)
		}
	})

	t.Run("large count", func(t *testing.T) {
		result := Yes("x", 1000)

		if len(result) != 1000 {
			t.Errorf("Yes() got %d items, want 1000", len(result))
		}
	})

	t.Run("unicode output", func(t *testing.T) {
		result := Yes("世界", 3)

		if len(result) != 3 {
			t.Errorf("Yes() got %d items, want 3", len(result))
		}

		for i, v := range result {
			if v != "世界" {
				t.Errorf("Yes()[%d] = %v, want '世界'", i, v)
			}
		}
	})

	t.Run("whitespace output", func(t *testing.T) {
		result := Yes("  ", 2)

		if len(result) != 2 {
			t.Errorf("Yes() got %d items, want 2", len(result))
		}

		for i, v := range result {
			if v != "  " {
				t.Errorf("Yes()[%d] = %v, want '  '", i, v)
			}
		}
	})

	t.Run("special characters", func(t *testing.T) {
		result := Yes("!@#$%", 2)

		if len(result) != 2 || result[0] != "!@#$%" {
			t.Errorf("Yes() special = %v", result)
		}
	})
}
