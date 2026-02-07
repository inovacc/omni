#!/usr/bin/env python3
"""Black-box tests for security commands (encrypt, decrypt, uuid, random, jwt)."""

import os
import sys
import re
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_exit_code, assert_regex


def main():
    print("=== Testing security commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # UUID tests
        @t.test("uuid_v4")
        def test_uuid_v4():
            result = t.run("uuid", "-v", "4")
            assert_exit_code(result, 0, "uuid v4")
            uuid_pattern = r'^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
            assert_regex(result.stdout.strip(), uuid_pattern, "uuid v4 format")

        @t.test("uuid_v7")
        def test_uuid_v7():
            result = t.run("uuid", "-v", "7")
            assert_exit_code(result, 0, "uuid v7")
            uuid_pattern = r'^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
            assert_regex(result.stdout.strip(), uuid_pattern, "uuid v7 format")

        @t.test("uuid_unique")
        def test_uuid_unique():
            r1 = t.run("uuid")
            r2 = t.run("uuid")
            if r1.stdout.strip() == r2.stdout.strip():
                raise AssertionError("UUIDs should be unique")

        @t.test("uuid_multiple")
        def test_uuid_multiple():
            result = t.run("uuid", "-n", "5")
            assert_exit_code(result, 0, "uuid -n 5")
            lines = result.stdout.strip().split('\n')
            if len(lines) != 5:
                raise AssertionError(f"expected 5 UUIDs, got {len(lines)}")

        # Random string
        @t.test("random_string")
        def test_random_string():
            result = t.run("random", "string", "-n", "32")
            assert_exit_code(result, 0, "random string")
            output = result.stdout.strip()
            if len(output) < 32:
                raise AssertionError(f"expected at least 32 chars, got {len(output)}")

        @t.test("random_hex")
        def test_random_hex():
            result = t.run("random", "hex", "-n", "16")
            assert_exit_code(result, 0, "random hex")
            output = result.stdout.strip()
            assert_regex(output, r'^[0-9a-f]+$', "random hex should be hex chars only")

        @t.test("random_int")
        def test_random_int():
            result = t.run("random", "int", "--min", "1", "--max", "100")
            assert_exit_code(result, 0, "random int")
            try:
                val = int(result.stdout.strip())
                if val < 1 or val > 100:
                    raise AssertionError(f"random int {val} not in range [1, 100]")
            except ValueError:
                raise AssertionError(f"random int output not a number: {result.stdout.strip()}")

        @t.test("random_password")
        def test_random_password():
            result = t.run("random", "password")
            assert_exit_code(result, 0, "random password")
            output = result.stdout.strip()
            if len(output) < 8:
                raise AssertionError(f"password should be at least 8 chars, got {len(output)}")

        # Encrypt/decrypt roundtrip
        @t.test("encrypt_decrypt_roundtrip")
        def test_encrypt_decrypt():
            plaintext = "secret message for testing"
            f = t.create_temp_file(plaintext, "plain.txt")
            enc_path = str(Path(t.temp_dir) / "encrypted.enc")
            dec_path = str(Path(t.temp_dir) / "decrypted.txt")

            # Encrypt
            result = t.run("encrypt", "-p", "mypassword", "-o", enc_path, str(f))
            assert_exit_code(result, 0, "encrypt")

            # Decrypt
            result = t.run("decrypt", "-p", "mypassword", "-o", dec_path, enc_path)
            assert_exit_code(result, 0, "decrypt")

            # Verify
            decrypted = Path(dec_path).read_text()
            assert_eq(decrypted, plaintext, "decrypted should match original")

        @t.test("encrypt_wrong_password")
        def test_encrypt_wrong_password():
            f = t.create_temp_file("secret", "secret.txt")
            enc_path = str(Path(t.temp_dir) / "wrongpw.enc")
            dec_path = str(Path(t.temp_dir) / "wrongpw_dec.txt")

            t.run("encrypt", "-p", "correct", "-o", enc_path, str(f))
            result = t.run("decrypt", "-p", "wrong", "-o", dec_path, enc_path)
            # Should fail or produce garbage
            if result.returncode == 0 and Path(dec_path).exists():
                content = Path(dec_path).read_text()
                if content == "secret":
                    raise AssertionError("wrong password should not decrypt correctly")

        # JWT decode
        @t.test("jwt_decode")
        def test_jwt_decode():
            # A standard test JWT token (expired, not a security risk)
            token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
            result = t.run("jwt", "decode", token)
            assert_exit_code(result, 0, "jwt decode")
            assert_contains(result.stdout, "John Doe", "jwt should contain name")

        # KSUID
        @t.test("ksuid_generate")
        def test_ksuid():
            result = t.run("ksuid")
            assert_exit_code(result, 0, "ksuid")
            output = result.stdout.strip()
            if len(output) < 20:
                raise AssertionError(f"KSUID too short: {output}")

        # ULID
        @t.test("ulid_generate")
        def test_ulid():
            result = t.run("ulid")
            assert_exit_code(result, 0, "ulid")
            output = result.stdout.strip()
            assert_regex(output, r'^[0-9A-Z]{26}$', "ULID format")

        # Nanoid
        @t.test("nanoid_generate")
        def test_nanoid():
            result = t.run("nanoid")
            assert_exit_code(result, 0, "nanoid")
            output = result.stdout.strip()
            if len(output) < 10:
                raise AssertionError(f"nanoid too short: {output}")

        # Run all tests
        test_uuid_v4()
        test_uuid_v7()
        test_uuid_unique()
        test_uuid_multiple()
        test_random_string()
        test_random_hex()
        test_random_int()
        test_random_password()
        test_encrypt_decrypt()
        test_encrypt_wrong_password()
        test_jwt_decode()
        test_ksuid()
        test_ulid()
        test_nanoid()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
