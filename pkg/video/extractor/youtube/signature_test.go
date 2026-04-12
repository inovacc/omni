package youtube

import (
	"testing"

	"github.com/inovacc/omni/pkg/video/types"
)

// ---- extractPlayerID ----

func TestExtractPlayerID(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://www.youtube.com/s/player/abc123def/player_ias.vflset/en_US/base.js", "abc123def"},
		{"https://www.youtube.com/s/player/xyz-ABC_1/base.js", "xyz-ABC_1"},
		{"/s/player/deadbeef/base.js", "deadbeef"},
		{"https://example.com/no-player-id", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := extractPlayerID(tt.url)
			if got != tt.want {
				t.Errorf("extractPlayerID(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

// ---- NewSignatureDecryptor ----

func TestNewSignatureDecryptor(t *testing.T) {
	d := NewSignatureDecryptor(nil)
	if d == nil {
		t.Fatal("NewSignatureDecryptor returned nil")
	}
}

// ---- ExtractPlayerURL ----

func TestExtractPlayerURL(t *testing.T) {
	d := NewSignatureDecryptor(nil)

	tests := []struct {
		name    string
		html    string
		wantURL string
	}{
		{
			name:    "jsUrl relative path",
			html:    `some html "jsUrl":"/s/player/abc/base.js" more html`,
			wantURL: "https://www.youtube.com/s/player/abc/base.js",
		},
		{
			name:    "PLAYER_JS_URL relative path",
			html:    `"PLAYER_JS_URL":"/s/player/def/base.js"`,
			wantURL: "https://www.youtube.com/s/player/def/base.js",
		},
		{
			name:    "protocol-relative URL",
			html:    `"jsUrl":"//www.youtube.com/s/player/ghi/base.js"`,
			wantURL: "https://www.youtube.com/s/player/ghi/base.js",
		},
		{
			name:    "absolute URL",
			html:    `"jsUrl":"https://www.youtube.com/s/player/jkl/base.js"`,
			wantURL: "https://www.youtube.com/s/player/jkl/base.js",
		},
		{
			name:    "no match",
			html:    `<html><body>no player url here</body></html>`,
			wantURL: "",
		},
		{
			name:    "empty HTML",
			html:    "",
			wantURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.ExtractPlayerURL(tt.html)
			if got != tt.wantURL {
				t.Errorf("ExtractPlayerURL = %q, want %q", got, tt.wantURL)
			}
		})
	}
}

// ---- findFunctionName ----

func TestFindFunctionName_NoMatch(t *testing.T) {
	d := &SignatureDecryptor{playerJS: "no function here"}
	got := d.findFunctionName(sigFuncNamePatterns)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestFindFunctionName_NsigNoMatch(t *testing.T) {
	d := &SignatureDecryptor{playerJS: "irrelevant js code"}
	got := d.findFunctionName(nsigFuncNamePatterns)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// ---- extractFunctionCode ----

func TestExtractFunctionCode_NotFound(t *testing.T) {
	d := &SignatureDecryptor{playerJS: "var x = 1;"}
	got := d.extractFunctionCode("nonExistentFn")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestExtractFunctionCode_VarForm(t *testing.T) {
	d := &SignatureDecryptor{
		playerJS: `var myFunc = function(a) { return a.split(""); };`,
	}
	got := d.extractFunctionCode("myFunc")
	if got == "" {
		t.Error("expected to extract function code for var form")
	}
}

func TestExtractFunctionCode_AssignmentForm(t *testing.T) {
	d := &SignatureDecryptor{
		playerJS: `myFunc = function(a) { return a.reverse(); };`,
	}
	got := d.extractFunctionCode("myFunc")
	if got == "" {
		t.Error("expected to extract function code for assignment form")
	}
}

func TestExtractFunctionCode_FunctionDecl(t *testing.T) {
	d := &SignatureDecryptor{
		playerJS: `function myFunc(a) { return a.join(""); }`,
	}
	got := d.extractFunctionCode("myFunc")
	if got == "" {
		t.Error("expected to extract function code for function declaration")
	}
}

// ---- addHelperObjects ----

func TestAddHelperObjects_NoHelpers(t *testing.T) {
	d := &SignatureDecryptor{playerJS: ""}
	code := "function fn(a) { return a; }"
	got := d.addHelperObjects(code)
	if got != code {
		t.Errorf("addHelperObjects with no helpers should return code unchanged, got %q", got)
	}
}

func TestAddHelperObjects_SkipsBuiltins(t *testing.T) {
	// String, Math, Array should not be looked up in playerJS.
	d := &SignatureDecryptor{playerJS: ""}
	code := `function fn(a) { String.fromCharCode(65); Math.floor(1.5); Array.isArray(a); }`
	got := d.addHelperObjects(code)
	if got != code {
		t.Errorf("addHelperObjects should skip builtins, got %q", got)
	}
}

// ---- buildNsigFunc fallback to identity ----

func TestBuildNsigFunc_FallbackIdentity(t *testing.T) {
	d := &SignatureDecryptor{playerJS: "// empty player"}
	d.buildNsigFunc()
	if d.nsigFunc == nil {
		t.Fatal("nsigFunc should be set to identity fallback")
	}
	result, err := d.nsigFunc("test_nsig")
	if err != nil {
		t.Errorf("identity nsigFunc should not error: %v", err)
	}
	if result != "test_nsig" {
		t.Errorf("identity nsigFunc should return input unchanged, got %q", result)
	}
}

// ---- decryptFormats (pure URL manipulation, no network) ----

func TestDecryptFormats_IdentityNsig(t *testing.T) {
	ext := &YoutubeExtractor{}

	d := &SignatureDecryptor{}
	d.buildNsigFunc() // sets identity function

	formats := []types.Format{
		{URL: "https://example.com/video?n=testNsig&expire=12345"},
		{URL: "https://example.com/video?other=param"},
		{URL: ""}, // empty URL → skip
	}

	result := ext.decryptFormats(formats, d)
	if len(result) != len(formats) {
		t.Errorf("decryptFormats: got %d formats, want %d", len(result), len(formats))
	}

	// With identity nsig, the n parameter should remain unchanged.
	if result[0].URL == "" {
		t.Error("first format URL should not be empty")
	}
}

func TestDecryptFormats_EmptyInput(t *testing.T) {
	ext := &YoutubeExtractor{}
	d := &SignatureDecryptor{}
	d.buildNsigFunc()

	result := ext.decryptFormats(nil, d)
	if len(result) != 0 {
		t.Errorf("expected 0 results for nil input, got %d", len(result))
	}
}

// ---- extractContinuationEntries ----

func TestExtractContinuationEntries_Empty(t *testing.T) {
	ext := &YoutubeChannelExtractor{}
	entries, token := ext.extractContinuationEntries(map[string]any{})
	if len(entries) != 0 || token != "" {
		t.Error("expected empty result for empty response")
	}
}

func TestExtractContinuationEntries_WithItems(t *testing.T) {
	ext := &YoutubeChannelExtractor{}

	resp := map[string]any{
		"onResponseReceivedActions": []any{
			map[string]any{
				"appendContinuationItemsAction": map[string]any{
					"continuationItems": []any{
						map[string]any{
							"richItemRenderer": map[string]any{
								"content": map[string]any{
									"videoRenderer": map[string]any{
										"videoId": "vid123",
										"title":   map[string]any{"simpleText": "Test Video"},
									},
								},
							},
						},
						map[string]any{
							"continuationItemRenderer": map[string]any{
								"continuationEndpoint": map[string]any{
									"continuationCommand": map[string]any{
										"token": "cont456",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	entries, token := ext.extractContinuationEntries(resp)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if token != "cont456" {
		t.Errorf("token = %q, want cont456", token)
	}
}

// ---- extractInitialEntries ----

func TestExtractInitialEntries_EmptyResponse(t *testing.T) {
	ext := &YoutubeChannelExtractor{}
	entries, token := ext.extractInitialEntries(map[string]any{})
	if len(entries) != 0 || token != "" {
		t.Error("expected empty result for empty response")
	}
}

func TestExtractInitialEntries_WithRichGrid(t *testing.T) {
	ext := &YoutubeChannelExtractor{}

	resp := map[string]any{
		"contents": map[string]any{
			"twoColumnBrowseResultsRenderer": map[string]any{
				"tabs": []any{
					map[string]any{
						"tabRenderer": map[string]any{
							"content": map[string]any{
								"richGridRenderer": map[string]any{
									"contents": []any{
										map[string]any{
											"richItemRenderer": map[string]any{
												"content": map[string]any{
													"videoRenderer": map[string]any{
														"videoId": "gridVid1",
														"title":   map[string]any{"simpleText": "Grid Video 1"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	entries, token := ext.extractInitialEntries(resp)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != "gridVid1" {
		t.Errorf("entry ID = %q, want gridVid1", entries[0].ID)
	}
	if token != "" {
		t.Errorf("expected no continuation token, got %q", token)
	}
}
