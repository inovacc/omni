#!/usr/bin/env python3
"""Black-box tests for encoding and hashing commands."""

import os
import sys
import re
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_exit_code, assert_regex


def main():
    print("=== Testing encoding commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # base64
        @t.test("base64_encode")
        def test_base64_encode():
            result = t.run("base64", stdin="hello world")
            assert_eq(result.stdout.strip(), "aGVsbG8gd29ybGQ=", "base64 encode")

        @t.test("base64_decode")
        def test_base64_decode():
            result = t.run("base64", "-d", stdin="aGVsbG8gd29ybGQ=")
            assert_eq(result.stdout.strip(), "hello world", "base64 decode")

        @t.test("base64_encode_file")
        def test_base64_encode_file():
            f = t.create_temp_file("test content", "b64.txt")
            result = t.run("base64", str(f))
            assert_exit_code(result, 0, "base64 file encode")
            assert_contains(result.stdout.strip(), "dGVzdCBjb250ZW50", "base64 file output")

        # base32
        @t.test("base32_encode")
        def test_base32_encode():
            result = t.run("base32", stdin="hello")
            assert_eq(result.stdout.strip(), "NBSWY3DP", "base32 encode")

        @t.test("base32_decode")
        def test_base32_decode():
            result = t.run("base32", "-d", stdin="NBSWY3DP")
            assert_eq(result.stdout.strip(), "hello", "base32 decode")

        # base58
        @t.test("base58_encode")
        def test_base58_encode():
            result = t.run("base58", stdin="hello")
            assert_exit_code(result, 0, "base58 encode")
            # base58 encoding varies, just check it produces output
            if not result.stdout.strip():
                raise AssertionError("base58 encode should produce output")

        # hex encoding
        @t.test("hex_encode")
        def test_hex_encode():
            result = t.run("hex", "encode", stdin="hello")
            assert_eq(result.stdout.strip(), "68656c6c6f", "hex encode")

        @t.test("hex_decode")
        def test_hex_decode():
            result = t.run("hex", "decode", stdin="68656c6c6f")
            assert_eq(result.stdout.strip(), "hello", "hex decode")

        # url encoding
        @t.test("url_encode")
        def test_url_encode():
            result = t.run("url", "encode", stdin="hello world&foo=bar")
            assert_exit_code(result, 0, "url encode")
            output = result.stdout.strip()
            if " " in output:
                raise AssertionError(f"url encode should escape spaces: {output}")

        @t.test("url_decode")
        def test_url_decode():
            result = t.run("url", "decode", stdin="hello%20world")
            assert_eq(result.stdout.strip(), "hello world", "url decode")

        # html encoding
        @t.test("html_encode")
        def test_html_encode():
            result = t.run("html", "encode", stdin="<div>test</div>")
            assert_exit_code(result, 0, "html encode")
            output = result.stdout.strip()
            assert_contains(output, "&lt;", "should escape <")

        @t.test("html_decode")
        def test_html_decode():
            result = t.run("html", "decode", stdin="&lt;div&gt;test&lt;/div&gt;")
            assert_eq(result.stdout.strip(), "<div>test</div>", "html decode")

        # md5sum
        @t.test("md5sum")
        def test_md5sum():
            f = t.create_temp_file("test\n", "md5.txt")
            result = t.run("md5sum", str(f))
            assert_exit_code(result, 0, "md5sum")
            assert_regex(result.stdout, r"^[0-9a-f]{32}\s", "md5 hash format")

        # sha256sum
        @t.test("sha256sum")
        def test_sha256sum():
            f = t.create_temp_file("test\n", "sha256.txt")
            result = t.run("sha256sum", str(f))
            assert_exit_code(result, 0, "sha256sum")
            assert_regex(result.stdout, r"^[0-9a-f]{64}\s", "sha256 hash format")

        # sha512sum
        @t.test("sha512sum")
        def test_sha512sum():
            f = t.create_temp_file("test\n", "sha512.txt")
            result = t.run("sha512sum", str(f))
            assert_exit_code(result, 0, "sha512sum")
            assert_regex(result.stdout, r"^[0-9a-f]{128}\s", "sha512 hash format")

        # crc32sum
        @t.test("crc32sum_file")
        def test_crc32sum_file():
            f = t.create_temp_file("test", "crc32.txt")
            result = t.run("crc32sum", str(f))
            assert_exit_code(result, 0, "crc32sum")
            assert_contains(result.stdout, "d87f7e0c", "crc32 of 'test'")

        @t.test("crc32sum_stdin")
        def test_crc32sum_stdin():
            result = t.run("crc32sum", stdin="test")
            assert_exit_code(result, 0, "crc32sum stdin")
            assert_contains(result.stdout, "d87f7e0c", "crc32 stdin")

        @t.test("crc32sum_format")
        def test_crc32sum_format():
            f = t.create_temp_file("hello", "crc32_fmt.txt")
            result = t.run("crc32sum", str(f))
            assert_exit_code(result, 0, "crc32sum format")
            assert_regex(result.stdout, r"^[0-9a-f]{8}\s", "crc32 hash is 8 hex chars")

        @t.test("crc32sum_multiple")
        def test_crc32sum_multiple():
            f1 = t.create_temp_file("aaa", "crc32_a.txt")
            f2 = t.create_temp_file("bbb", "crc32_b.txt")
            result = t.run("crc32sum", str(f1), str(f2))
            assert_exit_code(result, 0, "crc32sum multiple")
            lines = result.stdout.strip().split('\n')
            if len(lines) != 2:
                raise AssertionError(f"expected 2 lines, got {len(lines)}")

        # crc64sum
        @t.test("crc64sum_file")
        def test_crc64sum_file():
            f = t.create_temp_file("test", "crc64.txt")
            result = t.run("crc64sum", str(f))
            assert_exit_code(result, 0, "crc64sum")
            assert_regex(result.stdout, r"^[0-9a-f]{16}\s", "crc64 hash is 16 hex chars")

        @t.test("crc64sum_stdin")
        def test_crc64sum_stdin():
            result = t.run("crc64sum", stdin="test")
            assert_exit_code(result, 0, "crc64sum stdin")
            assert_regex(result.stdout, r"^[0-9a-f]{16}\s", "crc64 stdin format")

        @t.test("crc64sum_consistent")
        def test_crc64sum_consistent():
            f = t.create_temp_file("deterministic", "crc64_con.txt")
            r1 = t.run("crc64sum", str(f))
            r2 = t.run("crc64sum", str(f))
            assert_eq(r1.stdout, r2.stdout, "crc64 should be deterministic")

        @t.test("crc64sum_multiple")
        def test_crc64sum_multiple():
            f1 = t.create_temp_file("aaa", "crc64_a.txt")
            f2 = t.create_temp_file("bbb", "crc64_b.txt")
            result = t.run("crc64sum", str(f1), str(f2))
            assert_exit_code(result, 0, "crc64sum multiple")
            lines = result.stdout.strip().split('\n')
            if len(lines) != 2:
                raise AssertionError(f"expected 2 lines, got {len(lines)}")

        # hash -a crc32 / crc64
        @t.test("hash_crc32_algorithm")
        def test_hash_crc32_algorithm():
            f = t.create_temp_file("test", "hash_crc32.txt")
            result = t.run("hash", "-a", "crc32", str(f))
            assert_exit_code(result, 0, "hash -a crc32")
            assert_contains(result.stdout, "d87f7e0c", "hash -a crc32 value")

        @t.test("hash_crc64_algorithm")
        def test_hash_crc64_algorithm():
            f = t.create_temp_file("test", "hash_crc64.txt")
            result = t.run("hash", "-a", "crc64", str(f))
            assert_exit_code(result, 0, "hash -a crc64")
            assert_regex(result.stdout, r"^[0-9a-f]{16}\s", "hash -a crc64 format")

        # hash consistency
        @t.test("hash_consistent")
        def test_hash_consistent():
            f = t.create_temp_file("deterministic", "consistent.txt")
            r1 = t.run("sha256sum", str(f))
            r2 = t.run("sha256sum", str(f))
            assert_eq(r1.stdout, r2.stdout, "hashes should be deterministic")

        # xxd
        @t.test("xxd_basic")
        def test_xxd_basic():
            f = t.create_temp_file("hello", "xxd.txt")
            result = t.run("xxd", str(f))
            assert_exit_code(result, 0, "xxd")
            assert_contains(result.stdout, "6865 6c6c", "xxd hex dump")

        @t.test("xxd_reverse")
        def test_xxd_reverse():
            result = t.run("xxd", "-r", stdin="00000000: 6865 6c6c 6f                               hello")
            assert_exit_code(result, 0, "xxd reverse")

        @t.test("xxd_plain")
        def test_xxd_plain():
            f = t.create_temp_file("AB", "xxd_plain.txt")
            result = t.run("xxd", "-p", str(f))
            assert_exit_code(result, 0, "xxd plain")
            assert_contains(result.stdout, "4142", "xxd plain hex")

        # Run all tests
        test_base64_encode()
        test_base64_decode()
        test_base64_encode_file()
        test_base32_encode()
        test_base32_decode()
        test_base58_encode()
        test_hex_encode()
        test_hex_decode()
        test_url_encode()
        test_url_decode()
        test_html_encode()
        test_html_decode()
        test_md5sum()
        test_sha256sum()
        test_sha512sum()
        test_crc32sum_file()
        test_crc32sum_stdin()
        test_crc32sum_format()
        test_crc32sum_multiple()
        test_crc64sum_file()
        test_crc64sum_stdin()
        test_crc64sum_consistent()
        test_crc64sum_multiple()
        test_hash_crc32_algorithm()
        test_hash_crc64_algorithm()
        test_hash_consistent()
        test_xxd_basic()
        test_xxd_reverse()
        test_xxd_plain()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
