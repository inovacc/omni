package youtube

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/inovacc/omni/pkg/video/cache"
	"github.com/inovacc/omni/pkg/video/jsinterp"
	"github.com/inovacc/omni/pkg/video/nethttp"
)

var (
	playerURLRe  = regexp.MustCompile(`"jsUrl"\s*:\s*"([^"]+)"`)
	playerURLRe2 = regexp.MustCompile(`"PLAYER_JS_URL"\s*:\s*"([^"]+)"`)
	playerIDRe   = regexp.MustCompile(`/s/player/([a-zA-Z0-9_-]+)/`)

	// Patterns to find the signature decryption function name.
	sigFuncNamePatterns = []string{
		`\b[cs]\s*&&\s*[adf]\.set\([^,]+\s*,\s*encodeURIComponent\(([a-zA-Z0-9$]+)\(`,
		`\b[a-zA-Z0-9]+\s*&&\s*[a-zA-Z0-9]+\.set\([^,]+\s*,\s*encodeURIComponent\(([a-zA-Z0-9$]+)\(`,
		`\bm=([a-zA-Z0-9$]{2,})\(decodeURIComponent\(h\.s\)\)`,
		`\bc\s*&&\s*d\.set\([^,]+\s*,\s*(?:encodeURIComponent\s*\()([a-zA-Z0-9$]+)\(`,
		`\bc\s*&&\s*[a-z]\.set\([^,]+\s*,\s*([a-zA-Z0-9$]+)\(`,
		`\bc\s*&&\s*[a-z]\.set\([^,]+\s*,\s*encodeURIComponent\(([a-zA-Z0-9$]+)\(`,
	}

	// Patterns to find the nsig (throttle) function name.
	nsigFuncNamePatterns = []string{
		`\.get\("n"\)\)&&\(b=([a-zA-Z0-9$]+)(?:\[(\d+)\])?\([a-zA-Z0-9]\)`,
		`\b([a-zA-Z0-9$]+)\s*=\s*function\([a-zA-Z0-9]\)\s*\{var\s+b=a\.split\(""\)`,
		`([a-zA-Z0-9$]+)\s*=\s*function\(\s*a\s*\)\s*\{\s*a\s*=\s*a\.split\(\s*""\s*\)`,
	}
)

// SignatureDecryptor handles YouTube signature decryption using the player JS.
type SignatureDecryptor struct {
	cache     *cache.Cache
	playerURL string
	playerJS  string
	sigFunc   func(string) (string, error)
	nsigFunc  func(string) (string, error)
}

// NewSignatureDecryptor creates a new decryptor.
func NewSignatureDecryptor(c *cache.Cache) *SignatureDecryptor {
	return &SignatureDecryptor{cache: c}
}

// ExtractPlayerURL finds the player JS URL from a YouTube page.
func (s *SignatureDecryptor) ExtractPlayerURL(pageHTML string) string {
	for _, re := range []*regexp.Regexp{playerURLRe, playerURLRe2} {
		m := re.FindStringSubmatch(pageHTML)
		if m != nil {
			url := m[1]
			if strings.HasPrefix(url, "//") {
				url = "https:" + url
			} else if strings.HasPrefix(url, "/") {
				url = "https://www.youtube.com" + url
			}

			return url
		}
	}

	return ""
}

// LoadPlayer fetches and caches the player JavaScript.
func (s *SignatureDecryptor) LoadPlayer(ctx context.Context, client *nethttp.Client, playerURL string) error {
	s.playerURL = playerURL

	// Try cache first.
	playerID := extractPlayerID(playerURL)
	if playerID != "" && s.cache != nil {
		var cached string
		if s.cache.Load("player", playerID, &cached) && cached != "" {
			s.playerJS = cached
			return nil
		}
	}

	// Download player JS.
	body, err := client.GetString(ctx, playerURL)
	if err != nil {
		return fmt.Errorf("signature: fetching player JS: %w", err)
	}

	s.playerJS = body

	// Cache it.
	if playerID != "" && s.cache != nil {
		_ = s.cache.Store("player", playerID, body)
	}

	return nil
}

// DecryptSignature decrypts a scrambled signature.
func (s *SignatureDecryptor) DecryptSignature(sig string) (string, error) {
	if s.sigFunc == nil {
		if err := s.buildSigFunc(); err != nil {
			return "", err
		}
	}

	return s.sigFunc(sig)
}

// DecryptNsig decrypts the n parameter (throttling).
func (s *SignatureDecryptor) DecryptNsig(nsig string) (string, error) {
	if s.nsigFunc == nil {
		s.buildNsigFunc()
	}

	return s.nsigFunc(nsig)
}

func (s *SignatureDecryptor) buildSigFunc() error {
	funcName := s.findFunctionName(sigFuncNamePatterns)
	if funcName == "" {
		return fmt.Errorf("signature: could not find decryption function name")
	}

	// Extract the function and its dependencies from the player JS.
	funcCode := s.extractFunctionCode(funcName)
	if funcCode == "" {
		return fmt.Errorf("signature: could not extract function %s", funcName)
	}

	interp := jsinterp.New()

	fn, err := interp.ExtractFunction(funcCode, funcName)
	if err != nil {
		return fmt.Errorf("signature: %w", err)
	}

	s.sigFunc = fn

	return nil
}

func (s *SignatureDecryptor) buildNsigFunc() {
	identity := func(input string) (string, error) { return input, nil }

	funcName := s.findFunctionName(nsigFuncNamePatterns)
	if funcName == "" {
		// nsig decryption is optional — not all videos need it.
		s.nsigFunc = identity
		return
	}

	funcCode := s.extractFunctionCode(funcName)
	if funcCode == "" {
		s.nsigFunc = identity
		return
	}

	interp := jsinterp.New()

	fn, err := interp.ExtractFunction(funcCode, funcName)
	if err != nil {
		// Fall back to identity if nsig function can't be extracted.
		s.nsigFunc = identity
		return
	}

	s.nsigFunc = fn
}

func (s *SignatureDecryptor) findFunctionName(patterns []string) string {
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		m := re.FindStringSubmatch(s.playerJS)
		if len(m) > 1 {
			return m[1]
		}
	}

	return ""
}

func (s *SignatureDecryptor) extractFunctionCode(funcName string) string {
	// Escape special regex characters in function name.
	escaped := regexp.QuoteMeta(funcName)

	// Try to find: var funcName = function(a) { ... }
	patterns := []string{
		`(?s)(var\s+` + escaped + `\s*=\s*function\([^)]*\)\s*\{.+?\})\s*;`,
		`(?s)(` + escaped + `\s*=\s*function\([^)]*\)\s*\{.+?\})\s*;`,
		`(?s)(function\s+` + escaped + `\s*\([^)]*\)\s*\{.+?\})`,
	}

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		m := re.FindStringSubmatch(s.playerJS)
		if len(m) > 1 {
			code := m[1]
			// Also extract helper objects referenced by the function.
			code = s.addHelperObjects(code)

			return code
		}
	}

	return ""
}

func (s *SignatureDecryptor) addHelperObjects(code string) string {
	// Find references to helper objects like: xy.ab(a, 3)
	helperRe := regexp.MustCompile(`([a-zA-Z_$][a-zA-Z0-9_$]*)\.([a-zA-Z_$][a-zA-Z0-9_$]*)\(`)
	matches := helperRe.FindAllStringSubmatch(code, -1)

	seen := make(map[string]bool)

	var helpers []string

	for _, m := range matches {
		objName := m[1]
		if seen[objName] || objName == "String" || objName == "Math" || objName == "Array" {
			continue
		}

		seen[objName] = true

		// Extract the helper object definition.
		objPattern := `(?s)(var\s+` + regexp.QuoteMeta(objName) + `\s*=\s*\{.+?\}\s*\})\s*;`

		re, err := regexp.Compile(objPattern)
		if err != nil {
			continue
		}

		m := re.FindStringSubmatch(s.playerJS)
		if len(m) > 1 {
			helpers = append(helpers, m[1]+";")
		}
	}

	if len(helpers) > 0 {
		return strings.Join(helpers, "\n") + "\n" + code
	}

	return code
}

func extractPlayerID(url string) string {
	m := playerIDRe.FindStringSubmatch(url)
	if m != nil {
		return m[1]
	}

	return ""
}
