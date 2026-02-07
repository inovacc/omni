#!/usr/bin/env python3
"""Black-box tests for data processing commands (jq, yq, json, yaml, dotenv, toml, xml)."""

import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_exit_code, assert_json_valid


def main():
    print("=== Testing data processing commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # jq tests
        @t.test("jq_identity")
        def test_jq_identity():
            result = t.run("jq", ".", stdin='{"a":1}')
            assert_exit_code(result, 0, "jq identity")
            assert_json_valid(result.stdout, "jq output should be valid JSON")

        @t.test("jq_field")
        def test_jq_field():
            result = t.run("jq", ".name", stdin='{"name":"test","age":30}')
            assert_eq(result.stdout.strip(), '"test"', "jq field access")

        @t.test("jq_nested")
        def test_jq_nested():
            result = t.run("jq", ".user.name", stdin='{"user":{"name":"alice"}}')
            assert_eq(result.stdout.strip(), '"alice"', "jq nested field")

        @t.test("jq_array_index")
        def test_jq_array_index():
            result = t.run("jq", ".[1]", stdin='["a","b","c"]')
            assert_eq(result.stdout.strip(), '"b"', "jq array index")

        @t.test("jq_array_iterate")
        def test_jq_array_iterate():
            result = t.run("jq", ".[]", stdin='[1,2,3]')
            assert_exit_code(result, 0, "jq array iterate")

        @t.test("jq_keys")
        def test_jq_keys():
            result = t.run("jq", "keys", stdin='{"b":2,"a":1}')
            assert_exit_code(result, 0, "jq keys")

        @t.test("jq_length")
        def test_jq_length():
            result = t.run("jq", "length", stdin='[1,2,3]')
            assert_eq(result.stdout.strip(), "3", "jq length")

        @t.test("jq_raw")
        def test_jq_raw():
            result = t.run("jq", "-r", ".name", stdin='{"name":"test"}')
            assert_eq(result.stdout.strip(), "test", "jq raw output (no quotes)")

        @t.test("jq_compact")
        def test_jq_compact():
            result = t.run("jq", "-c", ".", stdin='{"a": 1, "b": 2}')
            assert_exit_code(result, 0, "jq compact")

        @t.test("jq_pipe")
        def test_jq_pipe():
            result = t.run("jq", ".[0] | .name", stdin='[{"name":"first"},{"name":"second"}]')
            assert_eq(result.stdout.strip(), '"first"', "jq pipe")

        @t.test("jq_type")
        def test_jq_type():
            result = t.run("jq", "type", stdin='"hello"')
            assert_eq(result.stdout.strip(), '"string"', "jq type")

        @t.test("jq_from_file")
        def test_jq_from_file():
            f = t.create_temp_file('{"key":"value"}', "data.json")
            result = t.run("jq", ".key", str(f))
            assert_eq(result.stdout.strip(), '"value"', "jq from file")

        # yq tests
        @t.test("yq_basic")
        def test_yq_basic():
            result = t.run("yq", ".name", stdin="name: test\nage: 30")
            assert_eq(result.stdout.strip(), "test", "yq basic field")

        @t.test("yq_nested")
        def test_yq_nested():
            result = t.run("yq", ".user.name", stdin="user:\n  name: alice")
            assert_eq(result.stdout.strip(), "alice", "yq nested field")

        @t.test("yq_array")
        def test_yq_array():
            result = t.run("yq", ".[0]", stdin="- a\n- b\n- c")
            assert_eq(result.stdout.strip(), "a", "yq array index")

        @t.test("yq_from_file")
        def test_yq_from_file():
            f = t.create_temp_file("key: value", "data.yaml")
            result = t.run("yq", ".key", str(f))
            assert_eq(result.stdout.strip(), "value", "yq from file")

        # json format/validate
        @t.test("json_fmt")
        def test_json_fmt():
            f = t.create_temp_file('{"a":1,"b":2}', "fmt.json")
            result = t.run("json", "fmt", str(f))
            assert_exit_code(result, 0, "json fmt")
            assert_contains(result.stdout, "\n", "json fmt should pretty-print")

        @t.test("json_minify")
        def test_json_minify():
            f = t.create_temp_file('{\n  "a": 1,\n  "b": 2\n}', "minify.json")
            result = t.run("json", "minify", str(f))
            assert_exit_code(result, 0, "json minify")

        @t.test("json_validate_valid")
        def test_json_validate_valid():
            f = t.create_temp_file('{"valid": true}', "valid.json")
            result = t.run("json", "validate", str(f))
            assert_exit_code(result, 0, "json validate valid")

        @t.test("json_validate_invalid")
        def test_json_validate_invalid():
            f = t.create_temp_file('{invalid json}', "invalid.json")
            result = t.run("json", "validate", str(f))
            # Should fail with non-zero exit or report error in output
            if result.returncode == 0 and "invalid" not in result.stderr.lower() and "invalid" not in result.stdout.lower():
                raise AssertionError("json validate should fail for invalid JSON")

        @t.test("json_tocsv")
        def test_json_tocsv():
            f = t.create_temp_file('[{"name":"a","val":1},{"name":"b","val":2}]', "tocsv.json")
            result = t.run("json", "tocsv", str(f))
            assert_exit_code(result, 0, "json tocsv")
            assert_contains(result.stdout, "name", "csv should have header")

        @t.test("json_tostruct")
        def test_json_tostruct():
            f = t.create_temp_file('{"name":"test","age":30}', "tostruct.json")
            result = t.run("json", "tostruct", str(f))
            assert_exit_code(result, 0, "json tostruct")

        # yaml validate
        @t.test("yaml_validate_valid")
        def test_yaml_validate_valid():
            result = t.run("yaml", "validate", stdin="name: test\nvalue: 123")
            assert_exit_code(result, 0, "yaml validate valid")

        @t.test("yaml_validate_invalid")
        def test_yaml_validate_invalid():
            result = t.run("yaml", "validate", stdin="name: \"unclosed")
            if result.returncode == 0:
                raise AssertionError("yaml validate should fail for invalid YAML")

        @t.test("yaml_tostruct")
        def test_yaml_tostruct():
            result = t.run("yaml", "tostruct", stdin="name: test\nage: 30")
            assert_exit_code(result, 0, "yaml tostruct")

        # dotenv
        @t.test("dotenv_parse")
        def test_dotenv_parse():
            f = t.create_temp_file("KEY=value\nNAME=test", "test.env")
            result = t.run("dotenv", str(f))
            assert_exit_code(result, 0, "dotenv parse")

        # toml validate
        @t.test("toml_validate_valid")
        def test_toml_validate_valid():
            f = t.create_temp_file('[server]\nhost = "localhost"\nport = 8080', "valid.toml")
            result = t.run("toml", "validate", str(f))
            assert_exit_code(result, 0, "toml validate valid")

        # Run all tests
        test_jq_identity()
        test_jq_field()
        test_jq_nested()
        test_jq_array_index()
        test_jq_array_iterate()
        test_jq_keys()
        test_jq_length()
        test_jq_raw()
        test_jq_compact()
        test_jq_pipe()
        test_jq_type()
        test_jq_from_file()
        test_yq_basic()
        test_yq_nested()
        test_yq_array()
        test_yq_from_file()
        test_json_fmt()
        test_json_minify()
        test_json_validate_valid()
        test_json_validate_invalid()
        test_json_tocsv()
        test_json_tostruct()
        test_yaml_validate_valid()
        test_yaml_validate_invalid()
        test_yaml_tostruct()
        test_dotenv_parse()
        test_toml_validate_valid()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
