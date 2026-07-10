<div id="top"></div>
<div align="center">
  <h1>LookThrough</h1>
  <h3>
    High-Performance Duplicate File Finder & Clean-up Utility in
    <a href="https://go.dev">Go</a> :rocket:
  </h3>
</div>

<details>
  <summary>
    <strong>Table of Contents</strong>
  </summary>
  <ol>
    <li><a href="#about">About</a></li>
    <li><a href="#how-it-works">How It Works</a></li>
    <li><a href="#features">Features</a></li>
    <li><a href="#requirements">Requirements</a></li>
    <li><a href="#installation">Installation</a></li>
    <li><a href="#usage">Usage</a>
      <ul>
        <li><a href="#1-using-cli-flags-recommended">1. Using CLI Flags</a></li>
        <li><a href="#2-passing-the-directory-directly">2. Passing the Directory Directly</a></li>
        <li><a href="#3-interactive-mode">3. Interactive Mode</a></li>
      </ul>
    </li>
    <li><a href="#system-exclusions">System Exclusions</a></li>
    <li><a href="#sample-output">Sample Output</a></li>
  </ol>
</details>

---

## About

**LookThrough** is a fast, multi-threaded command-line interface (CLI) tool built in Go. It is designed to recursively scan a target directory, identify exact duplicate files using cryptographic MD5 checksum validation, and isolate unique files into a structured output workspace. It automatically calculates name collisions and evaluates how much disk storage can be reclaimed.

---

## How It Works

1. **Pre-scan:** If the progress bar is enabled, the tool performs a rapid pre-scan to count the total number of valid files.
2. **Concurrent Hashing:** It walks through the directory tree, spawning concurrent worker pools (_goroutines_) to calculate MD5 hashes for files on the fly.
3. **Deduplication Strategy:**
   - **Unique File:** Copied directly to a new folder named `new-<target_directory>`.
   - **Identical Content (Duplicate Hash):** Skipped entirely to conserve space.
   - **Name Collision (Different Content):** If two different files share the exact same name, it safely appends a sequential numbering suffix (e.g., `filename(1).ext`) instead of overwriting data.
4. **Analysis Reporting:** Upon complete execution, the program compares total vs. unique counts and prints an optimized summary detailing total processed items and human-readable freed storage metrics.

---

## Features

- **High-Performance Processing:** Utilizes Go's internal scheduling with hardware-specific semaphore limits and $O(1)$ constant-time lookups for rapid duplicate detection during the processing loop.
- **Linear Time Verification:** Features an optimized hash-map validation technique to efficiently cross-verify data integrity in $O(N)$ linear time before finalizing.
- **RAM Throttling Protection:** Monitors memory overhead proactively via system checks, automatically pausing workers if RAM usage peaks to prevent out-of-memory (OOM) crashes.
- **Flexible Input Methods:** Supports both automated infrastructure scripting via command-line flags and human-centric interactive inputs.
- **Human-Readable Output:** Automatically converts byte-level storage sizes into human-readable strings (B, KB, MB, GB, TB) for better visibility.

---

## Requirements

- **Go Compiler:** Version 1.18 or higher.
- **Operating System:** Cross-platform support (Windows, macOS, and Linux distributions).

---

## Installation

Clone the repository and compile the highly optimized static binary using standard Go toolchains:

```sh
git clone https://github.com/LeonardoCG12/LookThrough.git
cd LookThrough
go build -o lookthrough main.go
```

## Usage

LookThrough offers three seamless execution paradigms to fit into manual testing or continuous delivery chains:

### 1. Using CLI Flags (Recommended)

Target a directory directly and initialize a visual feedback bar in your shell terminal interface using flags:

```sh
./lookthrough -p /absolute/path/to/target/directory -b
```

- `-p` : Explicit string pathway pointing directly towards your target directory folder.
- `-b` : Boolean flag instructing the script engine to render a visual console loading indicator.

### 2. Passing the Directory Directly

Execute the utility pass-through workflow by specifying a directory argument without flags:

```sh
./lookthrough /absolute/path/to/target/directory
```

### 3. Interactive Mode

Run the built application binary blindly without passing tracking arguments to enter an interactive step-by-step console routing wizard:

```sh
./lookthrough
Choose a directory to look through: /absolute/path/to/target/directory
```

## System Exclusions

To avoid cluttering your new consolidated repository directories, LookThrough natively catches and skips hidden OS ecosystem junk parameters automatically:

- `desktop.ini` (Windows Layout configuration metadata)
- `thumbs.db` (Windows Thumbnail cached database files)
- `.DS_Store` (macOS Desktop Services Store folder attributes)
- Prevents infinite loops by instantly ignoring any nested `new-*` generation directories encountered during secondary execution instances.

---

## Sample Output

When processing directory paths to completion successfully, a terminal diagnostic receipt is logged onto your environment:

```text
[+] SUCCESS
[+] ALL FILES HAVE BEEN COPIED

>>> Old Files: 1245
>>> New Files: 912
>>> Freed Storage: 4.2GB
```

If an empty folder or unsupported structure is matched:

```text
[-] FAIL
[-] NO FILES FOUND
```

<p align="right">[<a href="#top">Back to top</a>]</p>
