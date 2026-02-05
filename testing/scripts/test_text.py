#!/usr/bin/env python3
"""Black-box tests for text processing commands."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_not_contains, assert_line_count


def main():
    print("=== Testing text processing commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # grep tests
        @t.test("grep_basic")
        def test_grep_basic():
            f = t.create_temp_file("apple\nbanana\napricot\ncherry", "fruits.txt")
            result = t.run("grep", "ap", str(f))
            assert_contains(result.stdout, "apple", "grep should find apple")
            assert_contains(result.stdout, "apricot", "grep should find apricot")
            assert_not_contains(result.stdout, "banana", "grep should not find banana")

        @t.test("grep_case_insensitive")
        def test_grep_case_insensitive():
            f = t.create_temp_file("Apple\nBANANA\napricot", "mixed.txt")
            result = t.run("grep", "-i", "apple", str(f))
            assert_contains(result.stdout, "Apple", "grep -i should find Apple")

        @t.test("grep_invert")
        def test_grep_invert():
            f = t.create_temp_file("apple\nbanana\napricot", "fruits.txt")
            result = t.run("grep", "-v", "ap", str(f))
            assert_contains(result.stdout, "banana", "grep -v should find banana")
            assert_not_contains(result.stdout, "apple", "grep -v should not find apple")

        @t.test("grep_line_numbers")
        def test_grep_line_numbers():
            f = t.create_temp_file("apple\nbanana\napricot", "fruits.txt")
            result = t.run("grep", "-n", "banana", str(f))
            assert_contains(result.stdout, "2:", "grep -n should show line number")

        # sort tests
        @t.test("sort_basic")
        def test_sort_basic():
            f = t.create_temp_file("cherry\napple\nbanana", "unsorted.txt")
            result = t.run("sort", str(f))
            first_line = result.stdout.strip().split('\n')[0]
            assert_eq(first_line, "apple", "sort should put apple first")

        @t.test("sort_reverse")
        def test_sort_reverse():
            f = t.create_temp_file("apple\nbanana\ncherry", "sorted.txt")
            result = t.run("sort", "-r", str(f))
            first_line = result.stdout.strip().split('\n')[0]
            assert_eq(first_line, "cherry", "sort -r should put cherry first")

        @t.test("sort_numeric")
        def test_sort_numeric():
            f = t.create_temp_file("10\n2\n1\n20", "numbers.txt")
            result = t.run("sort", "-n", str(f))
            first_line = result.stdout.strip().split('\n')[0]
            assert_eq(first_line, "1", "sort -n should put 1 first")

        # uniq tests
        @t.test("uniq_basic")
        def test_uniq_basic():
            f = t.create_temp_file("apple\napple\nbanana\nbanana\nbanana\ncherry", "dups.txt")
            result = t.run("uniq", str(f))
            assert_line_count(result.stdout, 3, "uniq should return 3 unique lines")

        @t.test("uniq_count")
        def test_uniq_count():
            f = t.create_temp_file("apple\napple\nbanana", "dups.txt")
            result = t.run("uniq", "-c", str(f))
            assert_contains(result.stdout, "2", "uniq -c should show count 2 for apple")

        # wc tests
        @t.test("wc_lines")
        def test_wc_lines():
            f = t.create_temp_file("line1\nline2\nline3\n", "lines.txt")
            result = t.run("wc", "-l", str(f))
            assert_contains(result.stdout, "3", "wc -l should count 3 lines")

        @t.test("wc_words")
        def test_wc_words():
            f = t.create_temp_file("one two three four five", "words.txt")
            result = t.run("wc", "-w", str(f))
            assert_contains(result.stdout, "5", "wc -w should count 5 words")

        # tr tests
        @t.test("tr_lowercase")
        def test_tr_lowercase():
            result = t.run("tr", "[:upper:]", "[:lower:]", stdin="HELLO")
            assert_eq(result.stdout.strip(), "hello", "tr should convert to lowercase")

        @t.test("tr_delete")
        def test_tr_delete():
            result = t.run("tr", "-d", "0-9", stdin="hello123world")
            assert_eq(result.stdout.strip(), "helloworld", "tr -d should delete digits")

        # cut tests
        @t.test("cut_field")
        def test_cut_field():
            f = t.create_temp_file("a:b:c\n1:2:3", "delim.txt")
            result = t.run("cut", "-d", ":", "-f", "2", str(f))
            assert_contains(result.stdout, "b", "cut -f 2 should get second field")
            assert_contains(result.stdout, "2", "cut -f 2 should get second field")

        # Run all tests
        test_grep_basic()
        test_grep_case_insensitive()
        test_grep_invert()
        test_grep_line_numbers()
        test_sort_basic()
        test_sort_reverse()
        test_sort_numeric()
        test_uniq_basic()
        test_uniq_count()
        test_wc_lines()
        test_wc_words()
        test_tr_lowercase()
        test_tr_delete()
        test_cut_field()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
