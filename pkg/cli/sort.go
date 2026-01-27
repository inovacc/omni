package cli

func Pipeline(
	lines []string,
	grepPattern string,
	doSort bool,
	doUniq bool,
) []string {
	out := lines

	if grepPattern != "" {
		out = Grep(out, grepPattern)
	}

	if doSort {
		Sort(out)
	}

	if doUniq {
		out = Uniq(out)
	}

	return out
}
