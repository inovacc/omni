# Post-Build Testing & Benchmarks

This directory contains integration tests and benchmarks that run against the compiled `omni` binary.

## Overview

| Type | Purpose | Command |
|------|---------|---------|
| **Black-box Tests** | Verify CLI behavior and output | `task test:blackbox` |
| **Benchmarks** | Compare performance vs native Linux | `task test:benchmark` |

## Directory Structure

```
testing/
├── README.md           # This documentation
├── helpers.py          # Test utilities and assertions
├── run_all.py          # Main test runner
├── benchmark.py        # Performance benchmarks
└── scripts/
    ├── __init__.py
    ├── test_head_tail.py
    ├── test_text.py
    ├── test_file_ops.py
    └── test_utils.py
```

---

## Black-Box Tests

### Purpose

Unlike unit tests (`go test`), these tests:
- Test the actual compiled binary
- Verify end-to-end CLI behavior
- Test flag parsing and output formatting
- Catch integration issues between packages

### Running Tests

```bash
# Via Task (recommended)
task test:blackbox

# Manual execution
go build -o omni .
python testing/run_all.py

# Run specific test suite
python testing/scripts/test_head_tail.py

# Custom binary path
OMNI_BIN=./bin/omni python testing/run_all.py
```

### Test Suites

| Suite | Commands Tested | Test Count |
|-------|-----------------|------------|
| `test_head_tail.py` | head, tail, -NUM shortcuts | 9 |
| `test_text.py` | grep, sort, uniq, wc, tr, cut | 14 |
| `test_file_ops.py` | ls, cat, cp, mv, mkdir, rm, touch, stat, find | 14 |
| `test_utils.py` | echo, date, pwd, basename, dirname, seq, uuid, hash, base64, jq, yq | 17 |

### Test Cases Detail

#### head/tail Tests
| Test | Description |
|------|-------------|
| `head_default` | Verify default 10 lines output |
| `head_n_flag` | Test `-n NUM` flag |
| `head_numeric_shortcut` | Test `-NUM` shortcut (e.g., `-5`) |
| `head_bytes` | Test `-c NUM` bytes mode |
| `tail_default` | Verify default 10 lines output |
| `tail_n_flag` | Test `-n NUM` flag |
| `tail_numeric_shortcut` | Test `-NUM` shortcut |
| `tail_last_lines` | Verify correct last lines returned |

#### Text Processing Tests
| Test | Description |
|------|-------------|
| `grep_basic` | Pattern matching |
| `grep_case_insensitive` | `-i` flag |
| `grep_invert` | `-v` flag (invert match) |
| `grep_line_numbers` | `-n` flag |
| `sort_basic` | Alphabetical sort |
| `sort_reverse` | `-r` flag |
| `sort_numeric` | `-n` flag |
| `uniq_basic` | Remove consecutive duplicates |
| `uniq_count` | `-c` flag (count occurrences) |
| `wc_lines` | `-l` flag (line count) |
| `wc_words` | `-w` flag (word count) |
| `tr_lowercase` | Character translation |
| `tr_delete` | `-d` flag (delete chars) |
| `cut_field` | `-d` and `-f` flags |

#### File Operations Tests
| Test | Description |
|------|-------------|
| `ls_basic` | List directory |
| `ls_long` | `-l` long format |
| `ls_all` | `-a` show hidden |
| `cat_basic` | Output file content |
| `cat_multiple` | Concatenate files |
| `cat_line_numbers` | `-n` flag |
| `cp_basic` | Copy file |
| `mv_basic` | Move/rename file |
| `mkdir_basic` | Create directory |
| `mkdir_parents` | `-p` create parents |
| `rm_file` | Delete file |
| `touch_create` | Create/update timestamp |
| `stat_basic` | File information |
| `find_name` | Find by name pattern |

#### Utility Tests
| Test | Description |
|------|-------------|
| `echo_basic` | Print message |
| `echo_no_newline` | `-n` flag |
| `date_basic` | Current date/time |
| `date_format` | Custom format |
| `basename_*` | Extract filename |
| `dirname_*` | Extract directory |
| `seq_*` | Generate sequences |
| `uuid_basic` | Generate UUID |
| `hash_md5` | MD5 checksum |
| `hash_sha256` | SHA256 checksum |
| `base64_encode` | Base64 encode |
| `base64_decode` | Base64 decode |
| `jq_basic` | JSON query |
| `yq_basic` | YAML query |

### Writing New Tests

```python
#!/usr/bin/env python3
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_line_count

def main():
    print("=== Testing my command ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        @t.test("my_test_name")
        def test_my_feature():
            # Create test file if needed
            f = t.create_temp_file("content", "test.txt")

            # Run command
            result = t.run("mycommand", "-flag", str(f))

            # Assert results
            assert_eq(result.returncode, 0, "should succeed")
            assert_contains(result.stdout, "expected", "should contain expected")
            assert_line_count(result.stdout, 5, "should have 5 lines")

        # Execute test
        test_my_feature()

        t.print_summary()
        sys.exit(t.exit_code())

if __name__ == "__main__":
    main()
```

### Available Assertions

| Function | Description |
|----------|-------------|
| `assert_eq(actual, expected, msg)` | Exact equality |
| `assert_contains(haystack, needle, msg)` | Substring check |
| `assert_not_contains(haystack, needle, msg)` | Negative substring |
| `assert_exit_code(result, expected, msg)` | Command exit code |
| `assert_line_count(text, expected, msg)` | Number of lines |

---

## Benchmarks

### Purpose

Compare `omni` command performance against native Linux binaries:
- Identify performance bottlenecks
- Track performance regressions
- Validate optimization improvements

### Running Benchmarks

```bash
# Via Task
task test:benchmark

# Manual execution
go build -o omni .
python testing/benchmark.py

# Custom iterations
OMNI_BIN=./omni python -c "
from testing.benchmark import Benchmarker
# Custom benchmark setup
"
```

### Benchmark Cases

| Category | Benchmark | Description |
|----------|-----------|-------------|
| **head** | `head -n 10 (small)` | 10 lines from 3-line file |
| | `head -n 100 (large)` | 100 lines from 100k-line file |
| | `head -n 10000 (large)` | 10000 lines from 100k-line file |
| **tail** | `tail -n 10 (small)` | Last 10 lines (small file) |
| | `tail -n 100 (large)` | Last 100 lines (large file) |
| **cat** | `cat (small)` | 3-line file |
| | `cat (medium)` | 1000-line file |
| | `cat (large)` | 100k-line file |
| **wc** | `wc -l (small)` | Line count (small) |
| | `wc -l (large)` | Line count (large) |
| | `wc (large)` | Full word count |
| **grep** | `grep pattern` | Pattern search in 10k lines |
| | `grep -i pattern` | Case-insensitive search |
| | `grep -c count` | Count matches |
| **sort** | `sort (6k lines)` | Sort 6000 lines |
| | `sort -u unique` | Sort with unique |
| **uniq** | `uniq (sorted)` | Remove duplicates |
| | `uniq -c count` | Count occurrences |
| **tr** | `tr lowercase` | Character translation |
| **cut** | `cut -d, -f2` | Field extraction (10k lines) |
| **path** | `basename` | Extract filename |
| | `dirname` | Extract directory |
| **seq** | `seq 1000` | Generate 1000 numbers |
| | `seq 100000` | Generate 100k numbers |

### Benchmark Output

```
================================================================================
Benchmark                      Omni (ms)       Native (ms)     Ratio
================================================================================
head -n 10 (small)             5.23 ± 0.45     1.12 ± 0.08     4.67x slower
head -n 100 (large)            6.78 ± 0.52     2.34 ± 0.12     2.90x slower
cat (large)                    45.23 ± 2.34    12.56 ± 0.89    3.60x slower
grep pattern (10k lines)       23.45 ± 1.23    8.90 ± 0.45     2.64x slower
sort (6k lines)                12.34 ± 0.78    15.67 ± 0.92    1.27x faster
================================================================================
```

### Interpreting Results

- **Ratio < 1**: `omni` is faster than native
- **Ratio > 1**: `omni` is slower than native
- **± value**: Standard deviation across iterations

### Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `iterations` | 10 | Number of timed runs |
| `warmup` | 2 | Warmup runs (not counted) |

### Test Data Sizes

| File | Size | Purpose |
|------|------|---------|
| `small.txt` | 3 lines | Baseline overhead |
| `medium.txt` | 1,000 lines | Typical use case |
| `large.txt` | 100,000 lines | Stress test |
| `duplicates.txt` | 6,000 lines | sort/uniq tests |
| `grep.txt` | 10,000 lines | grep tests |
| `data.csv` | 10,000 lines | cut tests |

---

## CI Integration

Add to GitHub Actions workflow:

```yaml
- name: Run black-box tests
  run: |
    go build -o omni .
    python testing/run_all.py

- name: Run benchmarks (Linux only)
  if: runner.os == 'Linux'
  run: |
    python testing/benchmark.py
```

---

## Requirements

- Python 3.10+ (for type hints)
- Go toolchain (for building binary)
- Linux native commands (for benchmarks)
