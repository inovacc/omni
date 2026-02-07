package video

import (
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/omni/pkg/video"
	"github.com/inovacc/omni/pkg/video/downloader"
)

// MakeProgressFunc creates a progress callback that writes to w.
func MakeProgressFunc(w io.Writer, quiet bool) video.ProgressFunc {
	if quiet {
		return nil
	}

	return func(p video.ProgressInfo) {
		switch p.Status {
		case "downloading":
			pct := downloader.FormatPercent(p.DownloadedBytes, ptrOr(p.TotalBytes, 0))
			speed := downloader.FormatSpeed(p.Speed)
			eta := downloader.FormatETA(p.ETA)
			size := downloader.FormatBytes(p.DownloadedBytes)

			if p.FragmentIndex != nil && p.FragmentCount != nil {
				_, _ = fmt.Fprintf(w, "\r[download] %s of ~%s at %s ETA %s (frag %d/%d)",
					pct, size, speed, eta, *p.FragmentIndex+1, *p.FragmentCount)
			} else {
				total := "unknown"
				if p.TotalBytes != nil {
					total = downloader.FormatBytes(*p.TotalBytes)
				}

				_, _ = fmt.Fprintf(w, "\r[download] %s of %s at %s ETA %s",
					pct, total, speed, eta)
			}

		case "finished":
			size := downloader.FormatBytes(p.DownloadedBytes)
			_, _ = fmt.Fprintf(w, "\r[download] 100%% of %s in %.1fs\n", size, p.Elapsed)

		case "error":
			_, _ = fmt.Fprintf(w, "\n[download] Error: %s\n", p.Filename)
		}
	}
}

// FormatTable formats a list of formats as a human-readable table.
func FormatTable(formats []video.Format) string {
	var b strings.Builder

	// Header.
	_, _ = fmt.Fprintf(&b, "%-12s %-5s %-12s %-10s %-10s %-10s %-12s %s\n",
		"format_id", "ext", "resolution", "fps", "vcodec", "acodec", "filesize", "note")
	b.WriteString(strings.Repeat("-", 85))
	b.WriteString("\n")

	for _, f := range formats {
		res := f.FormatResolution()

		size := ""
		if fs := f.GetFilesize(); fs > 0 {
			size = downloader.FormatBytes(fs)
		}

		fps := ""
		if f.FPS != nil {
			fps = fmt.Sprintf("%.0f", *f.FPS)
		}

		vcodec := f.VCodec
		if vcodec == "" {
			vcodec = "-"
		}

		acodec := f.ACodec
		if acodec == "" {
			acodec = "-"
		}
		// Truncate long codec strings.
		if len(vcodec) > 10 {
			vcodec = vcodec[:10]
		}

		if len(acodec) > 10 {
			acodec = acodec[:10]
		}

		_, _ = fmt.Fprintf(&b, "%-12s %-5s %-12s %-10s %-10s %-10s %-12s %s\n",
			f.FormatID, f.Ext, res, fps, vcodec, acodec, size, f.FormatNote)
	}

	return b.String()
}

func ptrOr(p *int64, def int64) int64 {
	if p != nil {
		return *p
	}

	return def
}
