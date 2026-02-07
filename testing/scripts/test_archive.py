#!/usr/bin/env python3
"""Black-box tests for archive commands (tar, zip, gzip, bzip2, xz)."""

import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import OmniTester, assert_eq, assert_contains, assert_exit_code, assert_file_exists


def main():
    print("=== Testing archive commands ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        # Create test files
        f1 = t.create_temp_file("hello world\n", "file1.txt")
        f2 = t.create_temp_file("goodbye world\n", "file2.txt")
        f3 = t.create_temp_file("test data\n", "subdir/file3.txt")

        # tar create and list
        @t.test("tar_create")
        def test_tar_create():
            tar_path = Path(t.temp_dir) / "test.tar"
            result = t.run("tar", "-cf", str(tar_path), "-C", t.temp_dir, "file1.txt", "file2.txt")
            assert_exit_code(result, 0, "tar create should succeed")
            assert_file_exists(str(tar_path), "tar file should be created")

        @t.test("tar_list")
        def test_tar_list():
            tar_path = Path(t.temp_dir) / "test.tar"
            result = t.run("tar", "-tf", str(tar_path))
            assert_exit_code(result, 0, "tar list should succeed")
            assert_contains(result.stdout, "file1.txt", "tar list should show file1.txt")

        @t.test("tar_extract")
        def test_tar_extract():
            tar_path = Path(t.temp_dir) / "test.tar"
            extract_dir = Path(t.temp_dir) / "extracted"
            extract_dir.mkdir(exist_ok=True)
            result = t.run("tar", "-xf", str(tar_path), "-C", str(extract_dir))
            assert_exit_code(result, 0, "tar extract should succeed")
            assert_file_exists(str(extract_dir / "file1.txt"), "extracted file should exist")

        @t.test("tar_gz_create")
        def test_tar_gz_create():
            tar_path = Path(t.temp_dir) / "test.tar.gz"
            result = t.run("tar", "-czf", str(tar_path), "-C", t.temp_dir, "file1.txt", "file2.txt")
            assert_exit_code(result, 0, "tar -z create should succeed")
            assert_file_exists(str(tar_path), "tar.gz file should be created")

        @t.test("tar_gz_list")
        def test_tar_gz_list():
            tar_path = Path(t.temp_dir) / "test.tar.gz"
            result = t.run("tar", "-tzf", str(tar_path))
            assert_exit_code(result, 0, "tar -z list should succeed")
            assert_contains(result.stdout, "file1.txt", "tar.gz list should show file1.txt")

        # zip/unzip
        @t.test("zip_create")
        def test_zip_create():
            zip_path = Path(t.temp_dir) / "test.zip"
            result = t.run("zip", str(zip_path), str(f1), str(f2))
            assert_exit_code(result, 0, "zip create should succeed")
            assert_file_exists(str(zip_path), "zip file should be created")

        @t.test("unzip_list")
        def test_unzip_list():
            zip_path = Path(t.temp_dir) / "test.zip"
            result = t.run("unzip", "-l", str(zip_path))
            assert_exit_code(result, 0, "unzip list should succeed")

        @t.test("unzip_extract")
        def test_unzip_extract():
            zip_path = Path(t.temp_dir) / "test.zip"
            extract_dir = Path(t.temp_dir) / "unzipped"
            extract_dir.mkdir(exist_ok=True)
            result = t.run("unzip", str(zip_path), "-d", str(extract_dir))
            assert_exit_code(result, 0, "unzip extract should succeed")

        # gzip/gunzip
        @t.test("gzip_compress")
        def test_gzip_compress():
            gz_input = t.create_temp_file("compress me\n" * 100, "gzip_input.txt")
            result = t.run("gzip", str(gz_input))
            assert_exit_code(result, 0, "gzip should succeed")

        @t.test("gzip_stdout")
        def test_gzip_stdout():
            result = t.run("gzip", stdin="hello world\n" * 100)
            assert_exit_code(result, 0, "gzip from stdin should succeed")

        # bzip2
        @t.test("bzip2_compress")
        def test_bzip2_compress():
            bz_input = t.create_temp_file("compress me\n" * 100, "bzip2_input.txt")
            result = t.run("bzip2", str(bz_input))
            assert_exit_code(result, 0, "bzip2 should succeed")

        # xz
        @t.test("xz_compress")
        def test_xz_compress():
            xz_input = t.create_temp_file("compress me\n" * 100, "xz_input.txt")
            result = t.run("xz", str(xz_input))
            assert_exit_code(result, 0, "xz should succeed")

        # Run all tests
        test_tar_create()
        test_tar_list()
        test_tar_extract()
        test_tar_gz_create()
        test_tar_gz_list()
        test_zip_create()
        test_unzip_list()
        test_unzip_extract()
        test_gzip_compress()
        test_gzip_stdout()
        test_bzip2_compress()
        test_xz_compress()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
