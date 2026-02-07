#!/usr/bin/env python3
"""Benchmark omni commands against modern Rust-era tools (rg, fd, bat, etc.).

Gracefully skips benchmarks when native tools are not installed.
"""

import subprocess
import time
import statistics
import sys
import os
import tempfile
import shutil
from pathlib import Path
from dataclasses import dataclass, field
from typing import Optional


@dataclass
class BenchmarkResult:
    name: str
    omni_times: list[float]
    native_times: list[float]
    native_tool: str

    @property
    def omni_mean(self) -> float:
        return statistics.mean(self.omni_times) if self.omni_times else 0

    @property
    def native_mean(self) -> float:
        return statistics.mean(self.native_times) if self.native_times else 0

    @property
    def omni_stdev(self) -> float:
        return statistics.stdev(self.omni_times) if len(self.omni_times) > 1 else 0

    @property
    def native_stdev(self) -> float:
        return statistics.stdev(self.native_times) if len(self.native_times) > 1 else 0

    @property
    def ratio(self) -> float:
        if self.native_mean == 0:
            return float('inf')
        return self.omni_mean / self.native_mean

    @property
    def speedup(self) -> str:
        r = self.ratio
        if r < 1:
            return f"{1/r:.2f}x faster"
        elif r > 1:
            return f"{r:.2f}x slower"
        else:
            return "same"


class ModernBenchmarker:
    """Benchmark runner comparing omni vs modern tools (rg, fd, bat, etc.)."""

    def __init__(self, omni_bin: str = "./omni", iterations: int = 10, warmup: int = 2):
        self.omni_bin = omni_bin
        if os.name == "nt" and not self.omni_bin.endswith(".exe"):
            self.omni_bin += ".exe"
        self.iterations = iterations
        self.warmup = warmup
        self.temp_dir = tempfile.mkdtemp(prefix="omni_modern_bench_")
        self.results: list[BenchmarkResult] = []
        self.skipped: list[str] = []

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.cleanup()

    def cleanup(self):
        if os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)

    def has_tool(self, cmd: str) -> bool:
        return shutil.which(cmd) is not None

    def create_source_tree(self, dirs: int = 20, files_per_dir: int = 50,
                           lines_per_file: int = 100) -> Path:
        """Create a mock source tree with Go-like files."""
        root = Path(self.temp_dir) / "source_tree"
        root.mkdir(exist_ok=True)

        for d in range(dirs):
            pkg_dir = root / f"pkg{d:03d}"
            pkg_dir.mkdir(exist_ok=True)
            for f in range(files_per_dir):
                ext = ".go" if f % 3 != 0 else ".txt"
                filepath = pkg_dir / f"file{f:03d}{ext}"
                lines = []
                for line_num in range(lines_per_file):
                    if line_num % 10 == 0:
                        lines.append(f"// TODO: fix this issue in pkg{d:03d}/file{f:03d}")
                    elif line_num % 7 == 0:
                        lines.append(f'func handler{line_num}() error {{')
                    elif line_num % 7 == 1:
                        lines.append(f'    return fmt.Errorf("error in line {line_num}")')
                    else:
                        lines.append(f"    // line {line_num} of file {f} in dir {d}")
                filepath.write_text("\n".join(lines) + "\n")

        return root

    def create_large_file(self, lines: int, filename: str = "large.txt") -> Path:
        filepath = Path(self.temp_dir) / filename
        with open(filepath, 'w') as f:
            for i in range(lines):
                if i % 10 == 0:
                    f.write(f"ERROR: something went wrong at line {i}\n")
                else:
                    f.write(f"line {i}: {'x' * 80}\n")
        return filepath

    def create_json_file(self, entries: int, filename: str = "data.json") -> Path:
        filepath = Path(self.temp_dir) / filename
        import json
        data = [
            {"id": i, "name": f"item_{i}", "value": i * 1.5, "active": i % 2 == 0}
            for i in range(entries)
        ]
        filepath.write_text(json.dumps(data, indent=2))
        return filepath

    def create_deep_tree(self, depth: int = 8, breadth: int = 3) -> Path:
        root = Path(self.temp_dir) / "deep_tree"
        root.mkdir(exist_ok=True)

        def populate(current: Path, level: int):
            if level >= depth:
                return
            for i in range(breadth):
                child = current / f"dir_{level}_{i}"
                child.mkdir(exist_ok=True)
                for j in range(5):
                    (child / f"file_{j}.txt").write_text(f"content at depth {level}\n")
                populate(child, level + 1)

        populate(root, 0)
        return root

    def time_command(self, cmd: list[str], stdin: Optional[str] = None) -> float:
        start = time.perf_counter()
        subprocess.run(cmd, capture_output=True, text=True, input=stdin, timeout=30)
        return time.perf_counter() - start

    def benchmark(self, name: str, omni_cmd: list[str], native_cmd: list[str],
                  native_tool: str, stdin: Optional[str] = None):
        """Run benchmark comparing omni vs a modern tool."""
        # Warmup
        for _ in range(self.warmup):
            self.time_command(omni_cmd, stdin)
            self.time_command(native_cmd, stdin)

        omni_times = []
        native_times = []

        for _ in range(self.iterations):
            omni_times.append(self.time_command(omni_cmd, stdin))
            native_times.append(self.time_command(native_cmd, stdin))

        result = BenchmarkResult(
            name=name, omni_times=omni_times, native_times=native_times,
            native_tool=native_tool
        )
        self.results.append(result)
        return result

    def print_results(self):
        print()
        print("=" * 90)
        print(f"{'Benchmark':<35} {'Omni (ms)':<18} {'Native (ms)':<18} {'Comparison':<18}")
        print("=" * 90)

        for r in self.results:
            omni_str = f"{r.omni_mean*1000:.2f} ± {r.omni_stdev*1000:.2f}"
            native_str = f"{r.native_mean*1000:.2f} ± {r.native_stdev*1000:.2f}"
            print(f"{r.name:<35} {omni_str:<18} {native_str:<18} {r.speedup:<18}")

        print("=" * 90)
        print(f"Iterations: {self.iterations}, Warmup: {self.warmup}")

        if self.skipped:
            print(f"\nSkipped (tool not found): {', '.join(self.skipped)}")

        print()

    def print_summary(self):
        faster = sum(1 for r in self.results if r.ratio < 1)
        slower = sum(1 for r in self.results if r.ratio > 1)
        total = len(self.results)
        print(f"Summary: {total} benchmarks run, {faster} faster, {slower} slower, {total - faster - slower} tied")
        print(f"Skipped: {len(self.skipped)} (missing tools)")


def run_benchmarks():
    print("=" * 90)
    print("  Omni vs Modern Tools Benchmark (rg, fd, bat, jq, etc.)")
    print("=" * 90)
    print()

    omni_bin = os.environ.get("OMNI_BIN", "./omni")

    with ModernBenchmarker(omni_bin=omni_bin, iterations=10, warmup=2) as bench:
        # Create test fixtures
        print("Creating test fixtures...")
        source_tree = bench.create_source_tree(dirs=20, files_per_dir=50, lines_per_file=100)
        large_file = bench.create_large_file(100000, "large_search.txt")
        json_file = bench.create_json_file(10000, "bench_data.json")
        deep_tree = bench.create_deep_tree(depth=6, breadth=4)

        # Create file for compression benchmarks
        compress_file = bench.create_large_file(50000, "compress_data.txt")

        print(f"  Source tree: {source_tree} (20 dirs x 50 files x 100 lines)")
        print(f"  Large file:  {large_file} (100,000 lines)")
        print(f"  JSON file:   {json_file} (10,000 entries)")
        print(f"  Deep tree:   {deep_tree} (depth=6, breadth=4)")
        print()

        # === Search: omni rg vs rg (ripgrep) ===
        if bench.has_tool("rg"):
            print("Running search benchmarks (omni rg vs rg)...")

            bench.benchmark(
                "rg: pattern search",
                [omni_bin, "rg", "TODO", str(source_tree)],
                ["rg", "TODO", str(source_tree)],
                native_tool="rg"
            )

            bench.benchmark(
                "rg: case-insensitive",
                [omni_bin, "rg", "-i", "error", str(source_tree)],
                ["rg", "-i", "error", str(source_tree)],
                native_tool="rg"
            )

            bench.benchmark(
                "rg: fixed-string",
                [omni_bin, "rg", "-F", "fmt.Errorf", str(source_tree)],
                ["rg", "-F", "fmt.Errorf", str(source_tree)],
                native_tool="rg"
            )

            bench.benchmark(
                "rg: files-with-matches",
                [omni_bin, "rg", "-l", "handler", str(source_tree)],
                ["rg", "-l", "handler", str(source_tree)],
                native_tool="rg"
            )

            bench.benchmark(
                "rg: count matches",
                [omni_bin, "rg", "-c", "line", str(source_tree)],
                ["rg", "-c", "line", str(source_tree)],
                native_tool="rg"
            )

            bench.benchmark(
                "rg: large file search",
                [omni_bin, "rg", "ERROR", str(large_file)],
                ["rg", "ERROR", str(large_file)],
                native_tool="rg"
            )
        else:
            bench.skipped.append("rg (ripgrep)")

        # === Find: omni find vs fd ===
        if bench.has_tool("fd"):
            print("Running find benchmarks (omni find vs fd)...")

            bench.benchmark(
                "fd: name matching (*.go)",
                [omni_bin, "find", str(source_tree), "-name", "*.go"],
                ["fd", "-e", "go", str(source_tree)],
                native_tool="fd"
            )

            bench.benchmark(
                "fd: type filter (files)",
                [omni_bin, "find", str(source_tree), "-type", "f"],
                ["fd", "-t", "f", ".", str(source_tree)],
                native_tool="fd"
            )

            bench.benchmark(
                "fd: deep tree search",
                [omni_bin, "find", str(deep_tree), "-name", "*.txt"],
                ["fd", "-e", "txt", str(deep_tree)],
                native_tool="fd"
            )

            bench.benchmark(
                "fd: directories only",
                [omni_bin, "find", str(source_tree), "-type", "d"],
                ["fd", "-t", "d", ".", str(source_tree)],
                native_tool="fd"
            )
        else:
            bench.skipped.append("fd")

        # === Cat: omni cat vs bat ===
        if bench.has_tool("bat"):
            print("Running cat benchmarks (omni cat vs bat)...")

            bench.benchmark(
                "bat: large file",
                [omni_bin, "cat", str(large_file)],
                ["bat", "--plain", "--paging=never", str(large_file)],
                native_tool="bat"
            )

            bench.benchmark(
                "bat: line numbers",
                [omni_bin, "cat", "-n", str(large_file)],
                ["bat", "--plain", "--paging=never", "-n", str(large_file)],
                native_tool="bat"
            )

            # Multiple files
            files_in_dir = list((source_tree / "pkg000").glob("*.go"))[:10]
            file_args = [str(f) for f in files_in_dir]
            bench.benchmark(
                "bat: multiple files",
                [omni_bin, "cat"] + file_args,
                ["bat", "--plain", "--paging=never"] + file_args,
                native_tool="bat"
            )
        else:
            bench.skipped.append("bat")

        # === Tree: omni tree vs tree (native) ===
        if bench.has_tool("tree"):
            print("Running tree benchmarks (omni tree vs tree)...")

            bench.benchmark(
                "tree: shallow dir",
                [omni_bin, "tree", str(source_tree), "-L", "1"],
                ["tree", str(source_tree), "-L", "1"],
                native_tool="tree"
            )

            bench.benchmark(
                "tree: deep directory",
                [omni_bin, "tree", str(deep_tree)],
                ["tree", str(deep_tree)],
                native_tool="tree"
            )
        else:
            bench.skipped.append("tree")

        # === Hash: omni sha256sum vs sha256sum (native) ===
        if bench.has_tool("sha256sum"):
            print("Running hash benchmarks (omni vs native)...")

            bench.benchmark(
                "sha256sum: large file",
                [omni_bin, "sha256sum", str(large_file)],
                ["sha256sum", str(large_file)],
                native_tool="sha256sum"
            )

            if bench.has_tool("md5sum"):
                bench.benchmark(
                    "md5sum: large file",
                    [omni_bin, "md5sum", str(large_file)],
                    ["md5sum", str(large_file)],
                    native_tool="md5sum"
                )

            if bench.has_tool("sha512sum"):
                bench.benchmark(
                    "sha512sum: large file",
                    [omni_bin, "sha512sum", str(large_file)],
                    ["sha512sum", str(large_file)],
                    native_tool="sha512sum"
                )
        else:
            bench.skipped.append("sha256sum")

        # === JSON: omni jq vs jq (native) ===
        if bench.has_tool("jq"):
            print("Running JSON benchmarks (omni jq vs jq)...")

            bench.benchmark(
                "jq: simple query",
                [omni_bin, "jq", ".[0].name", str(json_file)],
                ["jq", ".[0].name", str(json_file)],
                native_tool="jq"
            )

            bench.benchmark(
                "jq: array iteration",
                [omni_bin, "jq", ".[].id", str(json_file)],
                ["jq", ".[].id", str(json_file)],
                native_tool="jq"
            )

            bench.benchmark(
                "jq: key extraction",
                [omni_bin, "jq", "[.[].name]", str(json_file)],
                ["jq", "[.[].name]", str(json_file)],
                native_tool="jq"
            )
        else:
            bench.skipped.append("jq")

        # === Compress: omni gzip vs gzip (native) ===
        if bench.has_tool("gzip"):
            print("Running compression benchmarks (omni vs native)...")

            # Compress
            omni_gz = str(Path(bench.temp_dir) / "omni_compressed.gz")
            native_gz = str(Path(bench.temp_dir) / "native_compressed.gz")

            bench.benchmark(
                "gzip: compress (stdout)",
                [omni_bin, "gzip", "-c", str(compress_file)],
                ["gzip", "-c", str(compress_file)],
                native_tool="gzip"
            )

            # Create gzipped file for decompression benchmark
            subprocess.run(
                ["gzip", "-c", str(compress_file)],
                capture_output=True
            )
            gz_path = Path(bench.temp_dir) / "decompress_test.gz"
            result = subprocess.run(
                ["gzip", "-c", str(compress_file)],
                capture_output=True
            )
            gz_path.write_bytes(result.stdout)

            if gz_path.exists() and gz_path.stat().st_size > 0:
                bench.benchmark(
                    "gzip: decompress (stdout)",
                    [omni_bin, "gunzip", "-c", str(gz_path)],
                    ["gunzip", "-c", str(gz_path)],
                    native_tool="gzip"
                )

            # tar create
            if bench.has_tool("tar"):
                omni_tar = str(Path(bench.temp_dir) / "omni_bench.tar.gz")
                native_tar = str(Path(bench.temp_dir) / "native_bench.tar.gz")
                tar_src = str(source_tree / "pkg000")

                bench.benchmark(
                    "tar: create .tar.gz",
                    [omni_bin, "tar", "-czf", omni_tar, tar_src],
                    ["tar", "-czf", native_tar, tar_src],
                    native_tool="tar"
                )

                # Create archive for extract benchmark
                subprocess.run(["tar", "-czf", native_tar, tar_src], capture_output=True)
                if Path(native_tar).exists():
                    omni_extract = str(Path(bench.temp_dir) / "omni_extract")
                    native_extract = str(Path(bench.temp_dir) / "native_extract")
                    os.makedirs(omni_extract, exist_ok=True)
                    os.makedirs(native_extract, exist_ok=True)

                    bench.benchmark(
                        "tar: extract .tar.gz",
                        [omni_bin, "tar", "-xzf", native_tar, "-C", omni_extract],
                        ["tar", "-xzf", native_tar, "-C", native_extract],
                        native_tool="tar"
                    )
        else:
            bench.skipped.append("gzip")

        # Print results
        bench.print_results()
        bench.print_summary()

        return 0


if __name__ == "__main__":
    sys.exit(run_benchmarks())
