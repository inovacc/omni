#!/usr/bin/env python3
"""Black-box tests for advanced text processing commands (sed, awk, tac, rev, paste, join, nl, fold, column, comm, cmp, strings, diff, split, shuf)."""

import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_exit_code, assert_line_count


def main():
    print("=== Testing advanced text commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # sed tests
        @t.test("sed_substitute")
        def test_sed_substitute():
            result = t.run("sed", "s/hello/world/", stdin="hello there\nhello again")
            assert_exit_code(result, 0, "sed substitute")
            assert_contains(result.stdout, "world there", "sed should substitute")

        @t.test("sed_global")
        def test_sed_global():
            result = t.run("sed", "s/a/b/g", stdin="banana")
            assert_eq(result.stdout.strip(), "bbnbnb", "sed global substitute")

        @t.test("sed_delete")
        def test_sed_delete():
            result = t.run("sed", "/delete/d", stdin="keep\ndelete\nkeep")
            assert_exit_code(result, 0, "sed delete")
            output = result.stdout.strip()
            if "delete" in output:
                raise AssertionError("sed /delete/d should remove matching lines")

        # awk tests
        @t.test("awk_print_field")
        def test_awk_field():
            result = t.run("awk", "{print $2}", stdin="one two three\nfour five six")
            assert_exit_code(result, 0, "awk print field")
            lines = result.stdout.strip().split('\n')
            assert_eq(lines[0], "two", "awk should print second field")

        @t.test("awk_pattern")
        def test_awk_pattern():
            result = t.run("awk", "/error/{print}", stdin="info: ok\nerror: fail\ninfo: ok")
            assert_exit_code(result, 0, "awk pattern")
            assert_contains(result.stdout, "error", "awk should match pattern")

        @t.test("awk_field_sep")
        def test_awk_field_sep():
            result = t.run("awk", "-F", ",", "{print $2}", stdin="a,b,c\nd,e,f")
            assert_exit_code(result, 0, "awk -F")
            assert_eq(result.stdout.strip().split('\n')[0], "b", "awk csv field")

        # tac (reverse lines)
        @t.test("tac_basic")
        def test_tac():
            result = t.run("tac", stdin="first\nsecond\nthird")
            assert_exit_code(result, 0, "tac")
            lines = result.stdout.strip().split('\n')
            assert_eq(lines[0], "third", "tac first line should be 'third'")
            assert_eq(lines[-1], "first", "tac last line should be 'first'")

        # rev (reverse characters)
        @t.test("rev_basic")
        def test_rev():
            result = t.run("rev", stdin="hello\nworld")
            assert_exit_code(result, 0, "rev")
            lines = result.stdout.strip().split('\n')
            assert_eq(lines[0], "olleh", "rev should reverse 'hello'")

        # nl (number lines)
        @t.test("nl_basic")
        def test_nl():
            result = t.run("nl", stdin="first\nsecond\nthird")
            assert_exit_code(result, 0, "nl")
            assert_contains(result.stdout, "1", "nl should number lines")
            assert_contains(result.stdout, "first", "nl should keep content")

        # fold (wrap lines)
        @t.test("fold_basic")
        def test_fold():
            long_line = "a" * 100
            result = t.run("fold", "-w", "20", stdin=long_line)
            assert_exit_code(result, 0, "fold")
            lines = result.stdout.strip().split('\n')
            if len(lines) < 5:
                raise AssertionError(f"fold -w 20 of 100-char line should produce 5+ lines, got {len(lines)}")

        # column
        @t.test("column_table")
        def test_column():
            result = t.run("column", "-t", stdin="name age\nalice 30\nbob 25")
            assert_exit_code(result, 0, "column -t")
            # Output should be aligned
            assert_contains(result.stdout, "name", "column should preserve content")

        # paste
        @t.test("paste_basic")
        def test_paste():
            f1 = t.create_temp_file("a\nb\nc", "paste1.txt")
            f2 = t.create_temp_file("1\n2\n3", "paste2.txt")
            result = t.run("paste", str(f1), str(f2))
            assert_exit_code(result, 0, "paste")
            assert_contains(result.stdout, "a\t1", "paste should join with tab")

        @t.test("paste_delimiter")
        def test_paste_delim():
            f1 = t.create_temp_file("a\nb\nc", "pasted1.txt")
            f2 = t.create_temp_file("1\n2\n3", "pasted2.txt")
            result = t.run("paste", "-d", ",", str(f1), str(f2))
            assert_exit_code(result, 0, "paste -d ,")
            assert_contains(result.stdout, "a,1", "paste should join with comma")

        # comm
        @t.test("comm_basic")
        def test_comm():
            f1 = t.create_temp_file("a\nb\nc", "comm1.txt")
            f2 = t.create_temp_file("b\nc\nd", "comm2.txt")
            result = t.run("comm", str(f1), str(f2))
            assert_exit_code(result, 0, "comm")

        # cmp
        @t.test("cmp_identical")
        def test_cmp_identical():
            f1 = t.create_temp_file("same", "cmp1.txt")
            f2 = t.create_temp_file("same", "cmp2.txt")
            result = t.run("cmp", str(f1), str(f2))
            assert_exit_code(result, 0, "cmp identical")

        @t.test("cmp_different")
        def test_cmp_different():
            f1 = t.create_temp_file("abc", "cmpd1.txt")
            f2 = t.create_temp_file("xyz", "cmpd2.txt")
            result = t.run("cmp", str(f1), str(f2))
            if result.returncode == 0:
                raise AssertionError("cmp should return non-zero for different files")

        # strings
        @t.test("strings_basic")
        def test_strings():
            # Create a file with mixed binary and text
            f = t.create_temp_file("hello\x00\x01\x02world\x00test", "strings.bin")
            result = t.run("strings", str(f))
            assert_exit_code(result, 0, "strings")
            assert_contains(result.stdout, "hello", "strings should find 'hello'")

        # diff
        @t.test("diff_basic")
        def test_diff():
            f1 = t.create_temp_file("line1\nline2\nline3", "diff1.txt")
            f2 = t.create_temp_file("line1\nchanged\nline3", "diff2.txt")
            result = t.run("diff", str(f1), str(f2))
            # diff returns non-zero when files differ, that's expected

        # shuf (shuffle lines)
        @t.test("shuf_basic")
        def test_shuf():
            result = t.run("shuf", stdin="1\n2\n3\n4\n5")
            assert_exit_code(result, 0, "shuf")
            assert_line_count(result.stdout, 5, "shuf should preserve line count")

        @t.test("shuf_n")
        def test_shuf_n():
            result = t.run("shuf", "-n", "3", stdin="1\n2\n3\n4\n5")
            assert_exit_code(result, 0, "shuf -n 3")
            assert_line_count(result.stdout, 3, "shuf -n 3 should output 3 lines")

        # join
        @t.test("join_basic")
        def test_join():
            f1 = t.create_temp_file("1 alice\n2 bob\n3 charlie", "join1.txt")
            f2 = t.create_temp_file("1 engineer\n2 designer\n3 manager", "join2.txt")
            result = t.run("join", str(f1), str(f2))
            assert_exit_code(result, 0, "join")
            assert_contains(result.stdout, "alice", "join should contain alice")
            assert_contains(result.stdout, "engineer", "join should contain engineer")

        # Run all tests
        test_sed_substitute()
        test_sed_global()
        test_sed_delete()
        test_awk_field()
        test_awk_pattern()
        test_awk_field_sep()
        test_tac()
        test_rev()
        test_nl()
        test_fold()
        test_column()
        test_paste()
        test_paste_delim()
        test_comm()
        test_cmp_identical()
        test_cmp_different()
        test_strings()
        test_diff()
        test_shuf()
        test_shuf_n()
        test_join()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
