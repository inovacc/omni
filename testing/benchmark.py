#!/usr/bin/env python3
"""Benchmark omni commands against native Linux binaries."""

import subprocess
import time
import statistics
import sys
import os
import tempfile
import shutil
from pathlib import Path
from dataclasses import dataclass
from typing import Callable, Optional


@dataclass
class BenchmarkResult:
    name: str
    omni_times: list[float]
    native_times: list[float]

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
        """Ratio of omni/native (< 1 means omni is faster)."""
        if self.native_mean == 0:
            return float('inf')
        return self.omni_mean / self.native_mean

    @property
    def speedup(self) -> str:
        """Human-readable speedup/slowdown."""
        r = self.ratio
        if r < 1:
            return f"{1/r:.2f}x faster"
        elif r > 1:
            return f"{r:.2f}x slower"
        else:
            return "same"


class Benchmarker:
    """Benchmark runner for comparing omni vs native commands."""

    def __init__(self, omni_bin: str = "./omni", iterations: int = 10, warmup: int = 2):
        self.omni_bin = omni_bin
        if os.name == "nt" and not self.omni_bin.endswith(".exe"):
            self.omni_bin += ".exe"
        self.iterations = iterations
        self.warmup = warmup
        self.temp_dir = tempfile.mkdtemp(prefix="omni_bench_")
        self.results: list[BenchmarkResult] = []

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.cleanup()

    def cleanup(self):
        if os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)

    def create_temp_file(self, content: str, filename: str = "bench.txt") -> Path:
        """Create a temporary file with given content."""
        filepath = Path(self.temp_dir) / filename
        filepath.write_text(content)
        return filepath

    def create_large_file(self, lines: int, filename: str = "large.txt") -> Path:
        """Create a large file with numbered lines."""
        filepath = Path(self.temp_dir) / filename
        with open(filepath, 'w') as f:
            for i in range(lines):
                f.write(f"line {i}: {'x' * 80}\n")
        return filepath

    def time_command(self, cmd: list[str], stdin: Optional[str] = None) -> float:
        """Time a single command execution."""
        start = time.perf_counter()
        subprocess.run(cmd, capture_output=True, text=True, input=stdin)
        return time.perf_counter() - start

    def benchmark(self, name: str, omni_cmd: list[str], native_cmd: list[str],
                  stdin: Optional[str] = None, setup: Optional[Callable] = None):
        """Run benchmark comparing omni vs native command."""
        if setup:
            setup()

        # Warmup
        for _ in range(self.warmup):
            self.time_command(omni_cmd, stdin)
            self.time_command(native_cmd, stdin)

        # Actual runs
        omni_times = []
        native_times = []

        for _ in range(self.iterations):
            if setup:
                setup()
            omni_times.append(self.time_command(omni_cmd, stdin))
            native_times.append(self.time_command(native_cmd, stdin))

        result = BenchmarkResult(name=name, omni_times=omni_times, native_times=native_times)
        self.results.append(result)
        return result

    def has_native(self, cmd: str) -> bool:
        """Check if native command is available."""
        return shutil.which(cmd) is not None

    def print_results(self):
        """Print benchmark results in a table."""
        print()
        print("=" * 80)
        print(f"{'Benchmark':<30} {'Omni (ms)':<15} {'Native (ms)':<15} {'Ratio':<15}")
        print("=" * 80)

        for r in self.results:
            omni_str = f"{r.omni_mean*1000:.2f} ± {r.omni_stdev*1000:.2f}"
            native_str = f"{r.native_mean*1000:.2f} ± {r.native_stdev*1000:.2f}" if r.native_mean > 0 else "N/A"
            ratio_str = r.speedup if r.native_mean > 0 else "N/A"
            print(f"{r.name:<30} {omni_str:<15} {native_str:<15} {ratio_str:<15}")

        print("=" * 80)
        print(f"Iterations: {self.iterations}, Warmup: {self.warmup}")
        print()


def run_benchmarks():
    """Run all benchmarks."""
    print("=" * 80)
    print("  Omni vs Native Linux Commands Benchmark")
    print("=" * 80)
    print()

    omni_bin = os.environ.get("OMNI_BIN", "./omni")

    with Benchmarker(omni_bin=omni_bin, iterations=10, warmup=2) as bench:
        # Create test files
        small_file = bench.create_temp_file("hello\nworld\ntest\n", "small.txt")
        medium_file = bench.create_large_file(1000, "medium.txt")
        large_file = bench.create_large_file(100000, "large.txt")

        # Create file with duplicates for sort/uniq
        dup_content = "\n".join(["apple", "banana", "apple", "cherry", "banana", "apple"] * 1000)
        dup_file = bench.create_temp_file(dup_content, "duplicates.txt")

        # Create file for grep
        grep_content = "\n".join([f"line {i}: {'error' if i % 10 == 0 else 'info'} message" for i in range(10000)])
        grep_file = bench.create_temp_file(grep_content, "grep.txt")

        print(f"Test files created in {bench.temp_dir}")
        print(f"  small.txt:  3 lines")
        print(f"  medium.txt: 1,000 lines")
        print(f"  large.txt:  100,000 lines")
        print()

        # head benchmarks
        print("Running head benchmarks...")
        if bench.has_native("head"):
            bench.benchmark(
                "head -n 10 (small)",
                [omni_bin, "head", "-n", "10", str(small_file)],
                ["head", "-n", "10", str(small_file)]
            )
            bench.benchmark(
                "head -n 100 (large)",
                [omni_bin, "head", "-n", "100", str(large_file)],
                ["head", "-n", "100", str(large_file)]
            )
            bench.benchmark(
                "head -n 10000 (large)",
                [omni_bin, "head", "-n", "10000", str(large_file)],
                ["head", "-n", "10000", str(large_file)]
            )

        # tail benchmarks
        print("Running tail benchmarks...")
        if bench.has_native("tail"):
            bench.benchmark(
                "tail -n 10 (small)",
                [omni_bin, "tail", "-n", "10", str(small_file)],
                ["tail", "-n", "10", str(small_file)]
            )
            bench.benchmark(
                "tail -n 100 (large)",
                [omni_bin, "tail", "-n", "100", str(large_file)],
                ["tail", "-n", "100", str(large_file)]
            )

        # cat benchmarks
        print("Running cat benchmarks...")
        if bench.has_native("cat"):
            bench.benchmark(
                "cat (small)",
                [omni_bin, "cat", str(small_file)],
                ["cat", str(small_file)]
            )
            bench.benchmark(
                "cat (medium)",
                [omni_bin, "cat", str(medium_file)],
                ["cat", str(medium_file)]
            )
            bench.benchmark(
                "cat (large)",
                [omni_bin, "cat", str(large_file)],
                ["cat", str(large_file)]
            )

        # wc benchmarks
        print("Running wc benchmarks...")
        if bench.has_native("wc"):
            bench.benchmark(
                "wc -l (small)",
                [omni_bin, "wc", "-l", str(small_file)],
                ["wc", "-l", str(small_file)]
            )
            bench.benchmark(
                "wc -l (large)",
                [omni_bin, "wc", "-l", str(large_file)],
                ["wc", "-l", str(large_file)]
            )
            bench.benchmark(
                "wc (large)",
                [omni_bin, "wc", str(large_file)],
                ["wc", str(large_file)]
            )

        # grep benchmarks
        print("Running grep benchmarks...")
        if bench.has_native("grep"):
            bench.benchmark(
                "grep pattern (10k lines)",
                [omni_bin, "grep", "error", str(grep_file)],
                ["grep", "error", str(grep_file)]
            )
            bench.benchmark(
                "grep -i pattern",
                [omni_bin, "grep", "-i", "ERROR", str(grep_file)],
                ["grep", "-i", "ERROR", str(grep_file)]
            )
            bench.benchmark(
                "grep -c count",
                [omni_bin, "grep", "-c", "error", str(grep_file)],
                ["grep", "-c", "error", str(grep_file)]
            )

        # sort benchmarks
        print("Running sort benchmarks...")
        if bench.has_native("sort"):
            bench.benchmark(
                "sort (6k lines)",
                [omni_bin, "sort", str(dup_file)],
                ["sort", str(dup_file)]
            )
            bench.benchmark(
                "sort -u unique",
                [omni_bin, "sort", "-u", str(dup_file)],
                ["sort", "-u", str(dup_file)]
            )

        # uniq benchmarks
        print("Running uniq benchmarks...")
        if bench.has_native("uniq"):
            # Need sorted input for uniq
            sorted_content = "\n".join(sorted(dup_content.split("\n")))
            sorted_file = bench.create_temp_file(sorted_content, "sorted.txt")
            bench.benchmark(
                "uniq (sorted input)",
                [omni_bin, "uniq", str(sorted_file)],
                ["uniq", str(sorted_file)]
            )
            bench.benchmark(
                "uniq -c count",
                [omni_bin, "uniq", "-c", str(sorted_file)],
                ["uniq", "-c", str(sorted_file)]
            )

        # tr benchmarks
        print("Running tr benchmarks...")
        if bench.has_native("tr"):
            large_content = large_file.read_text()
            bench.benchmark(
                "tr lowercase",
                [omni_bin, "tr", "[:upper:]", "[:lower:]"],
                ["tr", "[:upper:]", "[:lower:]"],
                stdin=large_content[:10000]
            )

        # cut benchmarks
        print("Running cut benchmarks...")
        if bench.has_native("cut"):
            csv_content = "\n".join([f"field1,field2,field3,field4,field5" for _ in range(10000)])
            csv_file = bench.create_temp_file(csv_content, "data.csv")
            bench.benchmark(
                "cut -d, -f2 (10k lines)",
                [omni_bin, "cut", "-d", ",", "-f", "2", str(csv_file)],
                ["cut", "-d", ",", "-f", "2", str(csv_file)]
            )

        # basename/dirname benchmarks
        print("Running path benchmarks...")
        if bench.has_native("basename"):
            bench.benchmark(
                "basename",
                [omni_bin, "basename", "/very/long/path/to/some/file.txt"],
                ["basename", "/very/long/path/to/some/file.txt"]
            )
        if bench.has_native("dirname"):
            bench.benchmark(
                "dirname",
                [omni_bin, "dirname", "/very/long/path/to/some/file.txt"],
                ["dirname", "/very/long/path/to/some/file.txt"]
            )

        # seq benchmarks
        print("Running seq benchmarks...")
        if bench.has_native("seq"):
            bench.benchmark(
                "seq 1000",
                [omni_bin, "seq", "1000"],
                ["seq", "1000"]
            )
            bench.benchmark(
                "seq 100000",
                [omni_bin, "seq", "100000"],
                ["seq", "100000"]
            )

        # Print results
        bench.print_results()

        # Summary
        faster = sum(1 for r in bench.results if r.ratio < 1)
        slower = sum(1 for r in bench.results if r.ratio > 1)
        same = sum(1 for r in bench.results if r.ratio == 1)
        print(f"Summary: {faster} faster, {slower} slower, {same} same")

        return 0 if slower <= faster else 1


if __name__ == "__main__":
    sys.exit(run_benchmarks())
