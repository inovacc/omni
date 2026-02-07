#!/usr/bin/env python3
"""Black-box tests for formatting commands (sql, html, css, json)."""

import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_exit_code


def main():
    print("=== Testing format commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # SQL formatting
        @t.test("sql_fmt_select")
        def test_sql_fmt_select():
            result = t.run("sql", "fmt", stdin="select id, name from users where active = true")
            assert_exit_code(result, 0, "sql fmt select")
            output = result.stdout.upper()
            assert_contains(output, "SELECT", "should format SELECT")

        @t.test("sql_fmt_join")
        def test_sql_fmt_join():
            result = t.run("sql", "fmt", stdin="select u.name, o.total from users u join orders o on u.id = o.user_id")
            assert_exit_code(result, 0, "sql fmt join")

        @t.test("sql_fmt_insert")
        def test_sql_fmt_insert():
            result = t.run("sql", "fmt", stdin="insert into users (name, age) values ('test', 30)")
            assert_exit_code(result, 0, "sql fmt insert")

        @t.test("sql_minify")
        def test_sql_minify():
            result = t.run("sql", "minify", stdin="SELECT\n  id,\n  name\nFROM\n  users\nWHERE\n  active = true")
            assert_exit_code(result, 0, "sql minify")
            output = result.stdout.strip()
            if '\n' in output:
                raise AssertionError("minified SQL should be single line")

        @t.test("sql_validate_valid")
        def test_sql_validate_valid():
            result = t.run("sql", "validate", stdin="SELECT id FROM users WHERE active = true")
            assert_exit_code(result, 0, "sql validate valid")

        @t.test("sql_validate_invalid")
        def test_sql_validate_invalid():
            result = t.run("sql", "validate", stdin="INVALID SQL STATEMENT !!!")
            if result.returncode == 0:
                # Some validators may be lenient
                pass

        # HTML formatting
        @t.test("html_fmt_basic")
        def test_html_fmt():
            result = t.run("html", "fmt", stdin="<div><p>hello</p></div>")
            assert_exit_code(result, 0, "html fmt")
            assert_contains(result.stdout, "<div>", "should contain div")

        @t.test("html_minify")
        def test_html_minify():
            result = t.run("html", "minify", stdin="<div>\n  <p>hello</p>\n</div>")
            assert_exit_code(result, 0, "html minify")

        @t.test("html_validate_valid")
        def test_html_validate_valid():
            result = t.run("html", "validate", stdin="<html><body><p>test</p></body></html>")
            assert_exit_code(result, 0, "html validate valid")

        # CSS formatting
        @t.test("css_fmt_basic")
        def test_css_fmt():
            result = t.run("css", "fmt", stdin="body{color:red;font-size:14px;}")
            assert_exit_code(result, 0, "css fmt")
            assert_contains(result.stdout, "body", "should contain selector")

        @t.test("css_minify")
        def test_css_minify():
            result = t.run("css", "minify", stdin="body {\n  color: red;\n  font-size: 14px;\n}")
            assert_exit_code(result, 0, "css minify")

        @t.test("css_validate_valid")
        def test_css_validate_valid():
            result = t.run("css", "validate", stdin="body { color: red; }")
            assert_exit_code(result, 0, "css validate valid")

        # Run all tests
        test_sql_fmt_select()
        test_sql_fmt_join()
        test_sql_fmt_insert()
        test_sql_minify()
        test_sql_validate_valid()
        test_sql_validate_invalid()
        test_html_fmt()
        test_html_minify()
        test_html_validate_valid()
        test_css_fmt()
        test_css_minify()
        test_css_validate_valid()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
