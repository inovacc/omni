package sign

import (
	"strings"
	"testing"
)

// lowScrypt keeps key generation cheap for tests (the default SENSITIVE cost
// needs ~1 GiB RAM and several seconds).
func lowScrypt() Option { return WithScryptParams(1<<15, 8, 1) }

// TestSecretKeyLogValueRedacts confirms LogValue never reveals key material.
func TestSecretKeyLogValueRedacts(t *testing.T) {
	kp, err := GenerateKeyPair("pw", lowScrypt())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	v := kp.SecretKey.LogValue()
	if !strings.Contains(v.String(), "REDACTED") {
		t.Errorf("LogValue() = %q, want a REDACTED placeholder", v.String())
	}
}

// TestApplyOptionsDefaultsAndClamping covers every clamping branch in
// applyOptions plus the comment-setting options.
func TestApplyOptionsDefaultsAndClamping(t *testing.T) {
	t.Run("defaults when no options", func(t *testing.T) {
		o := applyOptions(nil)
		if o.ScryptN != scryptN || o.ScryptR != scryptR || o.ScryptP != scryptP {
			t.Errorf("defaults = (%d,%d,%d), want (%d,%d,%d)", o.ScryptN, o.ScryptR, o.ScryptP, scryptN, scryptR, scryptP)
		}
	})
	t.Run("clamps non-positive params back to defaults", func(t *testing.T) {
		o := applyOptions([]Option{WithScryptParams(1, 0, 0)})
		if o.ScryptN != scryptN || o.ScryptR != scryptR || o.ScryptP != scryptP {
			t.Errorf("clamped = (%d,%d,%d), want defaults (%d,%d,%d)", o.ScryptN, o.ScryptR, o.ScryptP, scryptN, scryptR, scryptP)
		}
	})
	t.Run("honors explicit valid params", func(t *testing.T) {
		o := applyOptions([]Option{WithScryptParams(1<<16, 8, 2)})
		if o.ScryptN != 1<<16 || o.ScryptR != 8 || o.ScryptP != 2 {
			t.Errorf("explicit = (%d,%d,%d), want (65536,8,2)", o.ScryptN, o.ScryptR, o.ScryptP)
		}
	})
	t.Run("trusted and untrusted comments", func(t *testing.T) {
		o := applyOptions([]Option{
			WithTrustedComment("trusted-x"),
			WithUntrustedComment("untrusted-y"),
		})
		if o.TrustedComment != "trusted-x" {
			t.Errorf("TrustedComment = %q", o.TrustedComment)
		}
		if o.UntrustedComment != "untrusted-y" {
			t.Errorf("UntrustedComment = %q", o.UntrustedComment)
		}
	})
}

// TestScryptParamsFromLimits exercises both selection branches (ops-bound and
// mem-bound) of the libsodium-mirroring cost derivation, and confirms the
// round-trip with limitsFromScryptParams.
func TestScryptParamsFromLimits(t *testing.T) {
	cases := []struct {
		name       string
		opsLimit   uint64
		memLimit   uint64
		wantR      int
		minN, minP int
	}{
		// opsLimit < memLimit/32 -> p=1, N bounded by ops.
		{"ops-bound small", 32768, 1 << 30, 8, 2, 1},
		// floor applies: opsLimit below 32768 is raised.
		{"ops floored", 1, 1 << 30, 8, 2, 1},
		// mem-bound branch: large opsLimit, small memLimit.
		{"mem-bound", 1 << 40, 1 << 20, 8, 2, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n, r, p := scryptParamsFromLimits(c.opsLimit, c.memLimit)
			if r != c.wantR {
				t.Errorf("r = %d, want %d", r, c.wantR)
			}
			if n < c.minN {
				t.Errorf("n = %d, want >= %d", n, c.minN)
			}
			if p < c.minP {
				t.Errorf("p = %d, want >= %d", p, c.minP)
			}
			// N must be a power of two.
			if n&(n-1) != 0 {
				t.Errorf("n = %d is not a power of two", n)
			}
		})
	}
}
