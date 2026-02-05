#!/usr/bin/env python3
"""Black-box tests for head and tail commands."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_not_contains, assert_line_count


def main():
    print("=== Testing head/tail commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # Create test file
        test_content = "\n".join([f"line{i}" for i in range(1, 13)])
        test_file = t.create_temp_file(test_content, "lines.txt")

        @t.test("head_default")
        def test_head_default():
            result = t.run("head", str(test_file))
            assert_line_count(result.stdout, 10, "head default should return 10 lines")

        @t.test("head_n_flag")
        def test_head_n_flag():
            result = t.run("head", "-n", "5", str(test_file))
            assert_line_count(result.stdout, 5, "head -n 5 should return 5 lines")

        @t.test("head_numeric_shortcut")
        def test_head_numeric_shortcut():
            result = t.run("head", "-5", str(test_file))
            assert_line_count(result.stdout, 5, "head -5 shortcut should return 5 lines")

        @t.test("head_numeric_shortcut_content")
        def test_head_numeric_shortcut_content():
            result = t.run("head", "-3", str(test_file))
            assert_contains(result.stdout, "line1", "head -3 should contain line1")
            assert_contains(result.stdout, "line3", "head -3 should contain line3")
            assert_not_contains(result.stdout, "line4", "head -3 should not contain line4")

        @t.test("head_bytes")
        def test_head_bytes():
            result = t.run("head", "-c", "10", str(test_file))
            assert_eq(len(result.stdout), 10, "head -c 10 should return 10 bytes")

        @t.test("tail_default")
        def test_tail_default():
            result = t.run("tail", str(test_file))
            assert_line_count(result.stdout, 10, "tail default should return 10 lines")

        @t.test("tail_n_flag")
        def test_tail_n_flag():
            result = t.run("tail", "-n", "3", str(test_file))
            assert_line_count(result.stdout, 3, "tail -n 3 should return 3 lines")

        @t.test("tail_numeric_shortcut")
        def test_tail_numeric_shortcut():
            result = t.run("tail", "-3", str(test_file))
            assert_line_count(result.stdout, 3, "tail -3 shortcut should return 3 lines")

        @t.test("tail_last_lines")
        def test_tail_last_lines():
            result = t.run("tail", "-2", str(test_file))
            assert_contains(result.stdout, "line11", "tail -2 should contain line11")
            assert_contains(result.stdout, "line12", "tail -2 should contain line12")
            assert_not_contains(result.stdout, "line10", "tail -2 should not contain line10")

        # Run all tests
        test_head_default()
        test_head_n_flag()
        test_head_numeric_shortcut()
        test_head_numeric_shortcut_content()
        test_head_bytes()
        test_tail_default()
        test_tail_n_flag()
        test_tail_numeric_shortcut()
        test_tail_last_lines()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
