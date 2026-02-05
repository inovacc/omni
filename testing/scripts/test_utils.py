#!/usr/bin/env python3
"""Black-box tests for utility commands."""

import sys
import re
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_exit_code


def main():
    print("=== Testing utility commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # echo tests
        @t.test("echo_basic")
        def test_echo_basic():
            result = t.run("echo", "hello world")
            assert_eq(result.stdout.strip(), "hello world", "echo should print message")

        @t.test("echo_no_newline")
        def test_echo_no_newline():
            result = t.run("echo", "-n", "test")
            assert_eq(result.stdout, "test", "echo -n should not add newline")

        # date tests
        @t.test("date_basic")
        def test_date_basic():
            result = t.run("date")
            assert_exit_code(result, 0, "date should succeed")

        @t.test("date_format")
        def test_date_format():
            result = t.run("date", "+%Y")
            assert_contains(result.stdout, "202", "date +%Y should return year")

        # pwd tests
        @t.test("pwd_basic")
        def test_pwd_basic():
            result = t.run("pwd")
            assert_exit_code(result, 0, "pwd should succeed")

        # basename tests
        @t.test("basename_basic")
        def test_basename_basic():
            result = t.run("basename", "/path/to/file.txt")
            assert_eq(result.stdout.strip(), "file.txt", "basename should extract filename")

        @t.test("basename_suffix")
        def test_basename_suffix():
            result = t.run("basename", "/path/to/file.txt", ".txt")
            assert_eq(result.stdout.strip(), "file", "basename should strip suffix")

        # dirname tests
        @t.test("dirname_basic")
        def test_dirname_basic():
            result = t.run("dirname", "/path/to/file.txt")
            assert_eq(result.stdout.strip(), "/path/to", "dirname should extract directory")

        # seq tests
        @t.test("seq_basic")
        def test_seq_basic():
            result = t.run("seq", "5")
            lines = result.stdout.strip().split('\n')
            assert_eq(len(lines), 5, "seq 5 should output 5 lines")

        @t.test("seq_range")
        def test_seq_range():
            result = t.run("seq", "2", "5")
            first_line = result.stdout.strip().split('\n')[0]
            assert_eq(first_line, "2", "seq 2 5 should start at 2")

        # uuid tests
        @t.test("uuid_basic")
        def test_uuid_basic():
            result = t.run("uuid")
            uuid_pattern = r'^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
            if not re.match(uuid_pattern, result.stdout.strip(), re.IGNORECASE):
                raise AssertionError(f"Invalid UUID format: {result.stdout.strip()}")

        # hash tests
        @t.test("hash_md5")
        def test_hash_md5():
            f = t.create_temp_file("test content", "hash.txt")
            result = t.run("md5sum", str(f))
            assert_contains(result.stdout, "hash.txt", "md5sum should show filename")

        @t.test("hash_sha256")
        def test_hash_sha256():
            f = t.create_temp_file("test content", "hash.txt")
            result = t.run("sha256sum", str(f))
            assert_contains(result.stdout, "hash.txt", "sha256sum should show filename")

        # base64 tests
        @t.test("base64_encode")
        def test_base64_encode():
            result = t.run("base64", stdin="hello")
            assert_eq(result.stdout.strip(), "aGVsbG8=", "base64 should encode correctly")

        @t.test("base64_decode")
        def test_base64_decode():
            result = t.run("base64", "-d", stdin="aGVsbG8=")
            assert_eq(result.stdout.strip(), "hello", "base64 -d should decode correctly")

        # jq tests
        @t.test("jq_basic")
        def test_jq_basic():
            result = t.run("jq", ".name", stdin='{"name":"test"}')
            assert_eq(result.stdout.strip(), '"test"', "jq should extract field")

        # yq tests
        @t.test("yq_basic")
        def test_yq_basic():
            result = t.run("yq", ".name", stdin="name: test")
            assert_eq(result.stdout.strip(), "test", "yq should extract field")

        # Run all tests
        test_echo_basic()
        test_echo_no_newline()
        test_date_basic()
        test_date_format()
        test_pwd_basic()
        test_basename_basic()
        test_basename_suffix()
        test_dirname_basic()
        test_seq_basic()
        test_seq_range()
        test_uuid_basic()
        test_hash_md5()
        test_hash_sha256()
        test_base64_encode()
        test_base64_decode()
        test_jq_basic()
        test_yq_basic()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
