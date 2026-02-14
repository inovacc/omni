package youtube

import "fmt"

// InnerTube API client configurations.
// These mimic different YouTube client types to access video data.
// Client configs are sourced from yt-dlp's INNERTUBE_CLIENTS.

// ClientConfig holds the configuration for an InnerTube API client.
type ClientConfig struct {
	Name              string
	Version           string
	APIKey            string
	ClientName        int
	UserAgent         string
	Android           bool
	AndroidSDKVersion int
	DeviceMake        string
	DeviceModel       string
	OSName            string
	OSVersion         string
}

var (
	// android_vr — primary client, doesn't require PoToken.
	clientAndroidVR = ClientConfig{
		Name:              "ANDROID_VR",
		Version:           "1.71.26",
		ClientName:        28,
		UserAgent:         "com.google.android.apps.youtube.vr.oculus/1.71.26 (Linux; U; Android 12L; eureka-user Build/SQ3A.220605.009.A1) gzip",
		Android:           true,
		AndroidSDKVersion: 32,
		DeviceMake:        "Oculus",
		DeviceModel:       "Quest 3",
		OSName:            "Android",
		OSVersion:         "12L",
	}

	// web — requires PoToken for HTTPS/DASH streams but works for metadata.
	clientWeb = ClientConfig{
		Name:       "WEB",
		Version:    "2.20260114.08.00",
		ClientName: 1,
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
	}

	// web_safari — returns pre-merged HLS formats.
	clientWebSafari = ClientConfig{
		Name:       "WEB",
		Version:    "2.20260114.08.00",
		ClientName: 1,
		UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Safari/605.1.15,gzip(gfe)",
	}

	// web_embedded — sometimes works when other web clients fail.
	clientWebEmbedded = ClientConfig{
		Name:       "WEB_EMBEDDED_PLAYER",
		Version:    "1.20260115.01.00",
		ClientName: 56,
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
	}

	// android — needs PoToken for streams, good for metadata.
	clientAndroid = ClientConfig{
		Name:              "ANDROID",
		Version:           "21.02.35",
		ClientName:        3,
		UserAgent:         "com.google.android.youtube/21.02.35 (Linux; U; Android 11) gzip",
		Android:           true,
		AndroidSDKVersion: 30,
		OSName:            "Android",
		OSVersion:         "11",
	}

	// ios — HLS live streams, 60fps formats.
	clientIOS = ClientConfig{
		Name:       "IOS",
		Version:    "21.02.3",
		ClientName: 5,
		UserAgent:  "com.google.ios.youtube/21.02.3 (iPhone16,2; U; CPU iOS 18_3_2 like Mac OS X;)",
		DeviceMake: "Apple",
		DeviceModel: "iPhone16,2",
		OSName:     "iPhone",
		OSVersion:  "18.3.2.22D82",
	}

	// tv — Cobalt browser.
	clientTV = ClientConfig{
		Name:       "TVHTML5",
		Version:    "7.20260114.12.00",
		ClientName: 7,
		UserAgent:  "Mozilla/5.0 (ChromiumStylePlatform) Cobalt/25.lts.30.1034943-gold (unlike Gecko), Unknown_TV_Unknown_0/Unknown (Unknown, Unknown)",
	}

	// clientOrder defines the order to try clients.
	// android_vr is first as it doesn't need PoToken.
	clientOrder = []ClientConfig{
		clientAndroidVR,
		clientWebSafari,
		clientWeb,
		clientAndroid,
		clientIOS,
		clientWebEmbedded,
		clientTV,
	}
)

// InnerTubeRequest builds a request body for the InnerTube API.
func InnerTubeRequest(videoID string, cfg ClientConfig) map[string]any {
	clientCtx := map[string]any{
		"clientName":    cfg.Name,
		"clientVersion": cfg.Version,
		"hl":            "en",
	}

	if cfg.UserAgent != "" {
		clientCtx["userAgent"] = cfg.UserAgent
	}

	if cfg.Android && cfg.AndroidSDKVersion > 0 {
		clientCtx["androidSdkVersion"] = cfg.AndroidSDKVersion
	}

	if cfg.DeviceMake != "" {
		clientCtx["deviceMake"] = cfg.DeviceMake
	}

	if cfg.DeviceModel != "" {
		clientCtx["deviceModel"] = cfg.DeviceModel
	}

	if cfg.OSName != "" {
		clientCtx["osName"] = cfg.OSName
	}

	if cfg.OSVersion != "" {
		clientCtx["osVersion"] = cfg.OSVersion
	}

	body := map[string]any{
		"context": map[string]any{
			"client": clientCtx,
		},
		"videoId":        videoID,
		"contentCheckOk": true,
		"racyCheckOk":    true,
	}

	if cfg.Android {
		body["params"] = "CgIQBg=="
	}

	return body
}

// InnerTubePlayerURL returns the InnerTube player API endpoint.
func InnerTubePlayerURL(apiKey string) string {
	if apiKey != "" {
		return "https://www.youtube.com/youtubei/v1/player?key=" + apiKey + "&prettyPrint=false"
	}

	return "https://www.youtube.com/youtubei/v1/player?prettyPrint=false"
}

// InnerTubeSearchURL returns the InnerTube search API endpoint.
func InnerTubeSearchURL(apiKey string) string {
	if apiKey != "" {
		return "https://www.youtube.com/youtubei/v1/search?key=" + apiKey + "&prettyPrint=false"
	}

	return "https://www.youtube.com/youtubei/v1/search?prettyPrint=false"
}

// InnerTubeBrowseURL returns the InnerTube browse API endpoint.
func InnerTubeBrowseURL(apiKey string) string {
	if apiKey != "" {
		return "https://www.youtube.com/youtubei/v1/browse?key=" + apiKey + "&prettyPrint=false"
	}

	return "https://www.youtube.com/youtubei/v1/browse?prettyPrint=false"
}

// InnerTubeHeaders returns additional headers to set for InnerTube requests.
// If sapisidHash is non-empty, adds Authorization and X-Goog-AuthUser headers
// for authenticated requests (required for WEB client with cookies).
func InnerTubeHeaders(videoID string, cfg ClientConfig, sapisidHash string) map[string]string {
	h := map[string]string{
		"Origin":  "https://www.youtube.com",
		"Referer": "https://www.youtube.com/watch?v=" + videoID,
	}

	// Web clients send X-Youtube headers; mobile/embedded clients don't.
	if !cfg.Android {
		h["X-Youtube-Client-Name"] = fmt.Sprintf("%d", cfg.ClientName)
		h["X-Youtube-Client-Version"] = cfg.Version
	}

	if sapisidHash != "" {
		h["Authorization"] = sapisidHash
		h["X-Goog-AuthUser"] = "0"
	}

	return h
}
