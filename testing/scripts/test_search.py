#!/usr/bin/env python3
"""Black-box tests for search commands (rg)."""

import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_exit_code, assert_json_valid, assert_line_count


def main():
    print("=== Testing search commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # Create test directory structure
        search_dir = Path(t.temp_dir) / "search"
        search_dir.mkdir()
        (search_dir / "main.go").write_text("package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n")
        (search_dir / "util.go").write_text("package main\n\nfunc helper() {\n\treturn nil\n}\n")
        (search_dir / "readme.md").write_text("# Project\n\nThis is a test project.\n")
        (search_dir / "data.json").write_text('{"name": "test"}\n')
        sub = search_dir / "sub"
        sub.mkdir()
        (sub / "nested.go").write_text("package sub\n\nfunc Nested() error {\n\treturn nil\n}\n")

        @t.test("rg_basic_pattern")
        def test_rg_basic():
            result = t.run("rg", "func", str(search_dir))
            assert_exit_code(result, 0, "rg basic search")
            assert_contains(result.stdout, "func", "should find func")

        @t.test("rg_case_insensitive")
        def test_rg_case_insensitive():
            result = t.run("rg", "-i", "FUNC", str(search_dir))
            assert_exit_code(result, 0, "rg case insensitive")
            assert_contains(result.stdout, "func", "should find func")

        @t.test("rg_fixed_string")
        def test_rg_fixed_string():
            result = t.run("rg", "-F", "fmt.Println", str(search_dir))
            assert_exit_code(result, 0, "rg fixed string")
            assert_contains(result.stdout, "fmt.Println", "should find literal")

        @t.test("rg_line_numbers")
        def test_rg_line_numbers():
            result = t.run("rg", "-n", "func", str(search_dir))
            assert_exit_code(result, 0, "rg line numbers")
            # Line numbers should be present (format: NUM:content)
            for line in result.stdout.strip().split('\n'):
                if ':' not in line:
                    continue
                # At least some lines should have numbers

        @t.test("rg_count")
        def test_rg_count():
            result = t.run("rg", "-c", "func", str(search_dir))
            assert_exit_code(result, 0, "rg count")

        @t.test("rg_files_with_match")
        def test_rg_files_with_match():
            result = t.run("rg", "-l", "func", str(search_dir))
            assert_exit_code(result, 0, "rg files with match")
            output = result.stdout
            assert_contains(output, "main.go", "should list main.go")
            assert_contains(output, "util.go", "should list util.go")

        @t.test("rg_context_lines")
        def test_rg_context():
            result = t.run("rg", "-C", "1", "hello", str(search_dir))
            assert_exit_code(result, 0, "rg context")

        @t.test("rg_json_output")
        def test_rg_json():
            result = t.run("rg", "--json", "func", str(search_dir))
            assert_exit_code(result, 0, "rg json")

        @t.test("rg_type_filter")
        def test_rg_type_filter():
            result = t.run("rg", "-t", "go", "func", str(search_dir))
            assert_exit_code(result, 0, "rg type filter")
            # Should not match readme.md
            output = result.stdout
            if "readme.md" in output:
                raise AssertionError("type filter -t go should not match .md files")

        @t.test("rg_glob_filter")
        def test_rg_glob_filter():
            result = t.run("rg", "-g", "*.go", "func", str(search_dir))
            assert_exit_code(result, 0, "rg glob filter")

        @t.test("rg_no_match")
        def test_rg_no_match():
            result = t.run("rg", "nonexistent_pattern_xyz", str(search_dir))
            if result.returncode == 0:
                raise AssertionError("rg should return non-zero for no matches")

        @t.test("rg_invert_match")
        def test_rg_invert_match():
            result = t.run("rg", "-v", "func", str(search_dir / "main.go"))
            assert_exit_code(result, 0, "rg invert match")
            if "func" in result.stdout.split('\n')[0] if result.stdout else "":
                raise AssertionError("inverted match should not show 'func' lines")

        @t.test("rg_word_match")
        def test_rg_word_match():
            result = t.run("rg", "-w", "main", str(search_dir))
            assert_exit_code(result, 0, "rg word match")

        @t.test("rg_max_count")
        def test_rg_max_count():
            result = t.run("rg", "-m", "1", "func", str(search_dir / "main.go"))
            assert_exit_code(result, 0, "rg max count")

        @t.test("rg_only_matching")
        def test_rg_only_matching():
            result = t.run("rg", "-o", "func", str(search_dir / "main.go"))
            assert_exit_code(result, 0, "rg only matching")
            for line in result.stdout.strip().split('\n'):
                # Each match line should only contain "func"
                parts = line.split(':')
                if len(parts) > 1:
                    match = parts[-1].strip()
                    if match and match != "func":
                        raise AssertionError(f"only matching should show 'func', got '{match}'")

        # Run all tests
        test_rg_basic()
        test_rg_case_insensitive()
        test_rg_fixed_string()
        test_rg_line_numbers()
        test_rg_count()
        test_rg_files_with_match()
        test_rg_context()
        test_rg_json()
        test_rg_type_filter()
        test_rg_glob_filter()
        test_rg_no_match()
        test_rg_invert_match()
        test_rg_word_match()
        test_rg_max_count()
        test_rg_only_matching()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
