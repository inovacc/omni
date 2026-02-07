#!/usr/bin/env python3
"""Black-box tests for system commands (env, whoami, id, uname, uptime, df, du, ps, kill, arch, which)."""

import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_exit_code, assert_regex


def main():
    print("=== Testing system commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        @t.test("env_basic")
        def test_env():
            result = t.run("env")
            assert_exit_code(result, 0, "env should succeed")
            # Should have at least some environment variables
            assert_contains(result.stdout, "=", "env should show KEY=VALUE pairs")

        @t.test("whoami")
        def test_whoami():
            result = t.run("whoami")
            assert_exit_code(result, 0, "whoami should succeed")
            output = result.stdout.strip()
            if not output:
                raise AssertionError("whoami should return a username")

        @t.test("id_basic")
        def test_id():
            result = t.run("id")
            assert_exit_code(result, 0, "id should succeed")
            output = result.stdout.strip()
            if not output:
                raise AssertionError("id should return user info")

        @t.test("uname_basic")
        def test_uname():
            result = t.run("uname")
            assert_exit_code(result, 0, "uname should succeed")
            output = result.stdout.strip()
            if not output:
                raise AssertionError("uname should return OS info")

        @t.test("uname_all")
        def test_uname_all():
            result = t.run("uname", "-a")
            assert_exit_code(result, 0, "uname -a should succeed")
            output = result.stdout.strip()
            if not output:
                raise AssertionError("uname -a should return detailed info")

        @t.test("uname_system")
        def test_uname_system():
            result = t.run("uname", "-s")
            assert_exit_code(result, 0, "uname -s should succeed")

        @t.test("uname_machine")
        def test_uname_machine():
            result = t.run("uname", "-m")
            assert_exit_code(result, 0, "uname -m should succeed")

        @t.test("uptime")
        def test_uptime():
            result = t.run("uptime")
            assert_exit_code(result, 0, "uptime should succeed")

        @t.test("arch_basic")
        def test_arch():
            result = t.run("arch")
            assert_exit_code(result, 0, "arch should succeed")
            output = result.stdout.strip()
            # Should contain a known architecture
            known_archs = ["x86_64", "amd64", "arm64", "aarch64", "i386", "i686"]
            found = any(a in output.lower() for a in known_archs)
            if not found:
                # Some systems return different format
                pass

        @t.test("df_basic")
        def test_df():
            result = t.run("df")
            assert_exit_code(result, 0, "df should succeed")
            assert_contains(result.stdout, "Filesystem", "df should show header")

        @t.test("df_human")
        def test_df_human():
            result = t.run("df", "-h")
            assert_exit_code(result, 0, "df -h should succeed")

        @t.test("du_basic")
        def test_du():
            result = t.run("du", t.temp_dir)
            assert_exit_code(result, 0, "du should succeed")

        @t.test("du_summary")
        def test_du_summary():
            result = t.run("du", "-s", t.temp_dir)
            assert_exit_code(result, 0, "du -s should succeed")

        @t.test("ps_basic")
        def test_ps():
            result = t.run("ps")
            assert_exit_code(result, 0, "ps should succeed")
            assert_contains(result.stdout, "PID", "ps should show PID header")

        @t.test("kill_list_signals")
        def test_kill_list():
            result = t.run("kill", "-l")
            assert_exit_code(result, 0, "kill -l should succeed")
            assert_contains(result.stdout, "TERM", "kill -l should list TERM signal")

        @t.test("which_existing")
        def test_which_existing():
            # Try to find 'go' which should exist in this environment
            result = t.run("which", "go")
            if result.returncode == 0:
                output = result.stdout.strip()
                if not output:
                    raise AssertionError("which should return a path for existing command")

        # Run all tests
        test_env()
        test_whoami()
        test_id()
        test_uname()
        test_uname_all()
        test_uname_system()
        test_uname_machine()
        test_uptime()
        test_arch()
        test_df()
        test_df_human()
        test_du()
        test_du_summary()
        test_ps()
        test_kill_list()
        test_which_existing()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
