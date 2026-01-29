package rev

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// RunRev reverses lines character by character
func RunRev(w io.Writer, args []string) error {
	if len(args) == 0 {
		return revReader(w, os.Stdin)
	}

	for _, path := range args {
		if path == "-" {
			if err := revReader(w, os.Stdin); err != nil {
				return err
			}

			continue
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("rev: %w", err)
		}

		err = revReader(w, f)
		_ = f.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func revReader(w io.Writer, r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		reversed := reverseString(line)
		_, _ = fmt.Fprintln(w, reversed)
	}

	return scanner.Err()
}

func reverseString(s string) string {
	runes := []rune(s)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}
