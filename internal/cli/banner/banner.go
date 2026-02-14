package banner

import (
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/omni/pkg/figlet"
)

// Options configures the banner command.
type Options struct {
	Font  string // -f: font name (default "standard")
	Width int    // -w: max width (0 = unlimited)
	List  bool   // -l: list available fonts
}

// RunBanner generates an ASCII art banner from text.
func RunBanner(w io.Writer, r io.Reader, args []string, opts Options) error {
	if opts.List {
		for _, name := range figlet.ListFonts() {
			_, _ = fmt.Fprintln(w, name)
		}

		return nil
	}

	text := strings.Join(args, " ")

	// Read from stdin if no args provided
	if text == "" {
		data, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("banner: read stdin: %w", err)
		}

		text = strings.TrimRight(string(data), "\n\r")
	}

	if text == "" {
		return fmt.Errorf("banner: no text provided")
	}

	fontName := opts.Font
	if fontName == "" {
		fontName = "standard"
	}

	var renderOpts []figlet.Option

	renderOpts = append(renderOpts, figlet.WithFont(fontName))
	if opts.Width > 0 {
		renderOpts = append(renderOpts, figlet.WithWidth(opts.Width))
	}

	result, err := figlet.Render(text, renderOpts...)
	if err != nil {
		return fmt.Errorf("banner: %w", err)
	}

	_, _ = fmt.Fprintln(w, result)

	return nil
}
