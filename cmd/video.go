package cmd

import (
	"fmt"
	"strings"

	"github.com/inovacc/omni/internal/cli/video"
	pkgvideo "github.com/inovacc/omni/pkg/video"
	"github.com/spf13/cobra"
)

var videoCmd = &cobra.Command{
	Use:   "video [URL]",
	Short: "Download videos from YouTube and other platforms",
	Long: `Video downloader supporting YouTube and other video platforms.

Subcommands:
  download      Download video(s) from URL
  info          Show video metadata as JSON
  list-formats  List available download formats
  search        Search YouTube
  interactive   Interactive menu (download/info/formats/nerd stats)
  extractors    List supported sites

Examples:
  omni video download "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video info "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video list-formats "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video download -f worst "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video search "golang tutorial"
  omni video interactive
  omni video "https://www.youtube.com/watch?v=dQw4w9WgXcQ" --interactive
  omni video extractors`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		interactive, _ := cmd.Flags().GetBool("interactive")
		if interactive {
			return runVideoInteractive(cmd, args)
		}

		if len(args) > 0 {
			return fmt.Errorf("video: got URL without action; use --interactive or a subcommand (download/info/list-formats)")
		}

		return cmd.Help()
	},
}

var videoDownloadCmd = &cobra.Command{
	Use:     "download <URL>",
	Aliases: []string{"dl", "get"},
	Short:   "Download video(s) from URL",
	Long: `Download video from a URL.

Flags:
  -f, --format=SPEC       Format selector (default "best")
  -o, --output=TEMPLATE   Output filename template
  -q, --quiet             Suppress progress output
  --no-progress           Disable progress bar
  --rate-limit=RATE       Rate limit (e.g., "1M", "500K")
  -R, --retries=N         Number of retries (default 3)
  -c, --continue          Resume partial downloads
  --no-part               Don't use .part files
  --cookies=FILE          Netscape cookie file
  --proxy=URL             HTTP/SOCKS proxy
  --write-info-json       Write .info.json file
  --write-subs            Write subtitle files
  --no-playlist           Download single video, not playlist
  --playlist-start=N      Start index (1-based)
  --playlist-end=N        End index
  -v, --verbose           Verbose output

Format selectors:
  best          Best quality with video+audio (default)
  worst         Worst quality with video+audio
  bestvideo     Best video-only stream
  bestaudio     Best audio-only stream
  FORMAT_ID     Specific format by ID
  best[height<=720]   Best format with height <= 720

Output template variables:
  %(id)s, %(title)s, %(ext)s, %(uploader)s, %(upload_date)s,
  %(channel)s, %(format_id)s, %(resolution)s

Examples:
  omni video download "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video dl -f worst "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video dl -o "%(title)s-%(format_id)s.%(ext)s" URL
  omni video dl --rate-limit 1M URL
  omni video dl -c URL           # resume partial download`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := video.Options{}
		opts.Format, _ = cmd.Flags().GetString("format")
		opts.Output, _ = cmd.Flags().GetString("output")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.NoProgress, _ = cmd.Flags().GetBool("no-progress")
		opts.RateLimit, _ = cmd.Flags().GetString("rate-limit")
		opts.Retries, _ = cmd.Flags().GetInt("retries")
		opts.Continue, _ = cmd.Flags().GetBool("continue")
		opts.NoPart, _ = cmd.Flags().GetBool("no-part")
		opts.CookieFile, _ = cmd.Flags().GetString("cookies")
		opts.Proxy, _ = cmd.Flags().GetString("proxy")
		opts.WriteInfoJSON, _ = cmd.Flags().GetBool("write-info-json")
		opts.WriteSubs, _ = cmd.Flags().GetBool("write-subs")
		opts.NoPlaylist, _ = cmd.Flags().GetBool("no-playlist")
		opts.PlaylistStart, _ = cmd.Flags().GetInt("playlist-start")
		opts.PlaylistEnd, _ = cmd.Flags().GetInt("playlist-end")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		return video.RunDownload(cmd.OutOrStdout(), args, opts)
	},
}

var videoInfoCmd = &cobra.Command{
	Use:   "info <URL>",
	Short: "Show video metadata as JSON",
	Long: `Extract and display video metadata in JSON format.

Examples:
  omni video info "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video info URL | jq '.title'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := video.Options{}
		opts.CookieFile, _ = cmd.Flags().GetString("cookies")
		opts.Proxy, _ = cmd.Flags().GetString("proxy")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		return video.RunInfo(cmd.OutOrStdout(), args, opts)
	},
}

var videoListFormatsCmd = &cobra.Command{
	Use:     "list-formats <URL>",
	Aliases: []string{"formats", "lf"},
	Short:   "List available download formats",
	Long: `List all available download formats for a video.

Examples:
  omni video list-formats "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video formats URL --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := video.Options{}
		opts.CookieFile, _ = cmd.Flags().GetString("cookies")
		opts.Proxy, _ = cmd.Flags().GetString("proxy")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		return video.RunListFormats(cmd.OutOrStdout(), args, opts)
	},
}

var videoSearchCmd = &cobra.Command{
	Use:   "search <QUERY>",
	Short: "Search YouTube for videos",
	Long: `Search YouTube and display results.

Examples:
  omni video search "golang tutorial"
  omni video search "how to cook pasta"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		opts := video.Options{}
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		return video.RunInfo(cmd.OutOrStdout(), []string{"ytsearch:" + query}, opts)
	},
}

var videoExtractorsCmd = &cobra.Command{
	Use:   "extractors",
	Short: "List all supported sites/extractors",
	Long: `List all registered video extractors.

Examples:
  omni video extractors`,
	RunE: func(cmd *cobra.Command, args []string) error {
		names := pkgvideo.ListExtractors()
		for _, name := range names {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), name)
		}
		return nil
	},
}

var videoInteractiveCmd = &cobra.Command{
	Use:     "interactive [URL]",
	Aliases: []string{"i"},
	Short:   "Interactive video menu (formats, info, nerd stats, extractors)",
	Long: `Start an interactive menu for video operations.

The menu lets you:
  - input or change URL
  - list formats
  - show metadata JSON
  - inspect nerd stats
  - list supported extractors
  - download with best or custom format selector

Examples:
  omni video interactive
  omni video interactive "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
  omni video "https://www.youtube.com/watch?v=dQw4w9WgXcQ" --interactive`,
	Args: cobra.MaximumNArgs(1),
	RunE: runVideoInteractive,
}

func init() {
	rootCmd.AddCommand(videoCmd)

	videoCmd.AddCommand(videoDownloadCmd)
	videoCmd.AddCommand(videoInfoCmd)
	videoCmd.AddCommand(videoListFormatsCmd)
	videoCmd.AddCommand(videoSearchCmd)
	videoCmd.AddCommand(videoInteractiveCmd)
	videoCmd.AddCommand(videoExtractorsCmd)

	videoCmd.Flags().Bool("interactive", false, "start interactive mode (same as 'video interactive')")
	videoCmd.Flags().String("cookies", "", "Netscape cookie file path for interactive mode")
	videoCmd.Flags().String("proxy", "", "HTTP/SOCKS proxy URL for interactive mode")
	videoCmd.Flags().BoolP("verbose", "v", false, "verbose output for interactive mode")

	// download flags
	videoDownloadCmd.Flags().StringP("format", "f", "best", "format selector")
	videoDownloadCmd.Flags().StringP("output", "o", "", "output filename template")
	videoDownloadCmd.Flags().BoolP("quiet", "q", false, "suppress output")
	videoDownloadCmd.Flags().Bool("no-progress", false, "disable progress bar")
	videoDownloadCmd.Flags().String("rate-limit", "", "rate limit (e.g., 1M, 500K)")
	videoDownloadCmd.Flags().IntP("retries", "R", 3, "number of retries")
	videoDownloadCmd.Flags().BoolP("continue", "c", false, "resume partial downloads")
	videoDownloadCmd.Flags().Bool("no-part", false, "don't use .part files")
	videoDownloadCmd.Flags().String("cookies", "", "Netscape cookie file path")
	videoDownloadCmd.Flags().String("proxy", "", "HTTP/SOCKS proxy URL")
	videoDownloadCmd.Flags().Bool("write-info-json", false, "write .info.json file")
	videoDownloadCmd.Flags().Bool("write-subs", false, "write subtitle files")
	videoDownloadCmd.Flags().Bool("no-playlist", false, "download single video only")
	videoDownloadCmd.Flags().Int("playlist-start", 0, "playlist start index (1-based)")
	videoDownloadCmd.Flags().Int("playlist-end", 0, "playlist end index")
	videoDownloadCmd.Flags().BoolP("verbose", "v", false, "verbose output")

	// info flags
	videoInfoCmd.Flags().String("cookies", "", "Netscape cookie file path")
	videoInfoCmd.Flags().String("proxy", "", "HTTP/SOCKS proxy URL")
	videoInfoCmd.Flags().BoolP("verbose", "v", false, "verbose output")

	// list-formats flags
	videoListFormatsCmd.Flags().String("cookies", "", "Netscape cookie file path")
	videoListFormatsCmd.Flags().String("proxy", "", "HTTP/SOCKS proxy URL")
	videoListFormatsCmd.Flags().Bool("json", false, "output as JSON")

	// search flags
	videoSearchCmd.Flags().BoolP("verbose", "v", false, "verbose output")

	// interactive flags
	videoInteractiveCmd.Flags().StringP("format", "f", "best", "format selector used by 'download (best)'")
	videoInteractiveCmd.Flags().StringP("output", "o", "", "output filename template")
	videoInteractiveCmd.Flags().BoolP("quiet", "q", false, "suppress output")
	videoInteractiveCmd.Flags().Bool("no-progress", false, "disable progress bar")
	videoInteractiveCmd.Flags().String("rate-limit", "", "rate limit (e.g., 1M, 500K)")
	videoInteractiveCmd.Flags().IntP("retries", "R", 3, "number of retries")
	videoInteractiveCmd.Flags().BoolP("continue", "c", false, "resume partial downloads")
	videoInteractiveCmd.Flags().Bool("no-part", false, "don't use .part files")
	videoInteractiveCmd.Flags().String("cookies", "", "Netscape cookie file path")
	videoInteractiveCmd.Flags().String("proxy", "", "HTTP/SOCKS proxy URL")
	videoInteractiveCmd.Flags().Bool("write-info-json", false, "write .info.json file")
	videoInteractiveCmd.Flags().Bool("write-subs", false, "write subtitle files")
	videoInteractiveCmd.Flags().Bool("no-playlist", false, "download single video only")
	videoInteractiveCmd.Flags().Int("playlist-start", 0, "playlist start index (1-based)")
	videoInteractiveCmd.Flags().Int("playlist-end", 0, "playlist end index")
	videoInteractiveCmd.Flags().BoolP("verbose", "v", false, "verbose output")
}

func runVideoInteractive(cmd *cobra.Command, args []string) error {
	opts := video.Options{}
	opts.Format = videoStringFlag(cmd, "format")
	opts.Output = videoStringFlag(cmd, "output")
	opts.Quiet = videoBoolFlag(cmd, "quiet")
	opts.NoProgress = videoBoolFlag(cmd, "no-progress")
	opts.RateLimit = videoStringFlag(cmd, "rate-limit")
	opts.Retries = videoIntFlag(cmd, "retries")
	opts.Continue = videoBoolFlag(cmd, "continue")
	opts.NoPart = videoBoolFlag(cmd, "no-part")
	opts.CookieFile = videoStringFlag(cmd, "cookies")
	opts.Proxy = videoStringFlag(cmd, "proxy")
	opts.WriteInfoJSON = videoBoolFlag(cmd, "write-info-json")
	opts.WriteSubs = videoBoolFlag(cmd, "write-subs")
	opts.NoPlaylist = videoBoolFlag(cmd, "no-playlist")
	opts.PlaylistStart = videoIntFlag(cmd, "playlist-start")
	opts.PlaylistEnd = videoIntFlag(cmd, "playlist-end")
	opts.Verbose = videoBoolFlag(cmd, "verbose")

	return video.RunInteractive(cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin(), args, opts)
}

func videoStringFlag(cmd *cobra.Command, name string) string {
	if cmd.Flags().Lookup(name) == nil {
		return ""
	}

	value, _ := cmd.Flags().GetString(name)

	return value
}

func videoBoolFlag(cmd *cobra.Command, name string) bool {
	if cmd.Flags().Lookup(name) == nil {
		return false
	}

	value, _ := cmd.Flags().GetBool(name)

	return value
}

func videoIntFlag(cmd *cobra.Command, name string) int {
	if cmd.Flags().Lookup(name) == nil {
		return 0
	}

	value, _ := cmd.Flags().GetInt(name)

	return value
}
