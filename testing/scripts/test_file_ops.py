#!/usr/bin/env python3
"""Black-box tests for file operation commands."""

import sys
import os
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_exit_code


def main():
    print("=== Testing file operation commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # ls tests
        @t.test("ls_basic")
        def test_ls_basic():
            result = t.run("ls", t.temp_dir)
            assert_exit_code(result, 0, "ls should succeed")

        @t.test("ls_long")
        def test_ls_long():
            t.create_temp_file("test", "ls_test.txt")
            result = t.run("ls", "-l", t.temp_dir)
            assert_contains(result.stdout, "ls_test.txt", "ls -l should show filename")

        @t.test("ls_all")
        def test_ls_all():
            result = t.run("ls", "-a", t.temp_dir)
            assert_contains(result.stdout, ".", "ls -a should show . entry")

        # cat tests
        @t.test("cat_basic")
        def test_cat_basic():
            f = t.create_temp_file("hello world", "cat_test.txt")
            result = t.run("cat", str(f))
            assert_eq(result.stdout.strip(), "hello world", "cat should output file content")

        @t.test("cat_multiple")
        def test_cat_multiple():
            f1 = t.create_temp_file("hello", "cat1.txt")
            f2 = t.create_temp_file("world", "cat2.txt")
            result = t.run("cat", str(f1), str(f2))
            assert_contains(result.stdout, "hello", "cat should include first file")
            assert_contains(result.stdout, "world", "cat should include second file")

        @t.test("cat_line_numbers")
        def test_cat_line_numbers():
            f = t.create_temp_file("line1\nline2\nline3", "numbered.txt")
            result = t.run("cat", "-n", str(f))
            assert_contains(result.stdout, "1", "cat -n should show line numbers")

        # cp tests
        @t.test("cp_basic")
        def test_cp_basic():
            src = t.create_temp_file("copy me", "src.txt")
            dst = Path(t.temp_dir) / "dst.txt"
            result = t.run("cp", str(src), str(dst))
            assert_exit_code(result, 0, "cp should succeed")
            assert_eq(dst.read_text(), "copy me", "cp should copy file content")

        # mv tests
        @t.test("mv_basic")
        def test_mv_basic():
            src = t.create_temp_file("move me", "move_src.txt")
            dst = Path(t.temp_dir) / "move_dst.txt"
            result = t.run("mv", str(src), str(dst))
            assert_exit_code(result, 0, "mv should succeed")
            assert not src.exists(), "mv should remove source file"
            assert_eq(dst.read_text(), "move me", "mv should move file content")

        # mkdir tests
        @t.test("mkdir_basic")
        def test_mkdir_basic():
            dir_path = Path(t.temp_dir) / "newdir"
            result = t.run("mkdir", str(dir_path))
            assert_exit_code(result, 0, "mkdir should succeed")
            assert dir_path.is_dir(), "mkdir should create directory"

        @t.test("mkdir_parents")
        def test_mkdir_parents():
            dir_path = Path(t.temp_dir) / "parent" / "child" / "grandchild"
            result = t.run("mkdir", "-p", str(dir_path))
            assert_exit_code(result, 0, "mkdir -p should succeed")
            assert dir_path.is_dir(), "mkdir -p should create nested directories"

        # rm tests
        @t.test("rm_file")
        def test_rm_file():
            f = t.create_temp_file("delete me", "delete.txt")
            result = t.run("rm", str(f))
            assert_exit_code(result, 0, "rm should succeed")
            assert not f.exists(), "rm should delete file"

        # touch tests
        @t.test("touch_create")
        def test_touch_create():
            f = Path(t.temp_dir) / "touched.txt"
            result = t.run("touch", str(f))
            assert_exit_code(result, 0, "touch should succeed")
            assert f.exists(), "touch should create file"

        # stat tests
        @t.test("stat_basic")
        def test_stat_basic():
            f = t.create_temp_file("stat me", "stat_test.txt")
            result = t.run("stat", str(f))
            assert_contains(result.stdout, "stat_test.txt", "stat should show filename")

        # find tests
        @t.test("find_name")
        def test_find_name():
            t.create_temp_file("test", "find_me.txt")
            t.create_temp_file("test", "skip.log")
            result = t.run("find", t.temp_dir, "-name", "*.txt")
            assert_contains(result.stdout, "find_me.txt", "find should locate .txt file")

        # Run all tests
        test_ls_basic()
        test_ls_long()
        test_ls_all()
        test_cat_basic()
        test_cat_multiple()
        test_cat_line_numbers()
        test_cp_basic()
        test_mv_basic()
        test_mkdir_basic()
        test_mkdir_parents()
        test_rm_file()
        test_touch_create()
        test_stat_basic()
        test_find_name()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
