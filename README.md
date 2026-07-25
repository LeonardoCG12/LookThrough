<div id="top"></div>

<div align="center">
  <img src="assets/branding/lookthrough_logo.png" alt="LookThrough Logo" width="350">
  <h3>
    🔎 High-Performance File Deduplication & Organization Utility in
    <a href="https://go.dev">Go</a> 🔍
  </h3>
</div>

<details>
  <summary>
    <strong>📑 Table of Contents</strong>
  </summary>
  <ol>
    <li><a href="#about">About</a></li>
    <li><a href="#how-it-works">How It Works</a></li>
    <li><a href="#features">Features</a></li>
    <li><a href="#requirements">Requirements</a></li>
    <li><a href="#installation">Installation</a></li>
    <li>
      <a href="#usage">Usage</a>
      <ul>
        <li><a href="#using-cli-flags">1. Using CLI Flags</a></li>
        <li><a href="#passing-the-directory-directly">2. Passing the Directory Directly</a></li>
        <li><a href="#interactive-mode">3. Interactive Mode</a></li>
      </ul>
    </li>
    <li><a href="#file-categorization-summary">File Categorization Summary</a></li>
    <li><a href="#system-exclusions">System Exclusions</a></li>
    <li><a href="#sample-output">Sample Output</a></li>
  </ol>
</details>

---

<div id="about"></div>

## ℹ️ About

**LookThrough** is a concurrent command-line interface (CLI) tool built in Go. It recursively scans a target directory, identifies content duplicates based on SHA-256 hashes, and copies one representative of each unique file into a structured output directory.

The output directory is created inside the target directory and is named `new-<target_directory>`. LookThrough resolves filename collisions without overwriting files and reports the total logical size of duplicate files that were not copied to the output.

LookThrough does not delete or modify the original source files.

---

<div id="how-it-works"></div>

## ⚙️ How It Works

1. **File Counting:** A first sequential `filepath.WalkDir` pass counts eligible source files so the progress indicator has a fixed final total before processing begins.
2. **Concurrent Processing:** A second `WalkDir` pass feeds files into a bounded worker pool. File metadata lookup, SHA-256 hashing, and copying run concurrently in the workers.
3. **Hash-Based Deduplication:**
   - **Unique Content:** Copied to `new-<target_directory>`.
   - **Duplicate Hash:** Skipped so the same content is not copied more than once.
   - **Name Collision with Different Content:** A sequential suffix such as `filename(1).ext` is added instead of overwriting an existing file.
4. **Optional Safe Copying:** With `-safe-copy`, each destination is first written to a temporary file in the final directory and renamed only after copying and closing succeed.
5. **Optional Categorization:** With `-s`, unique files are placed into category directories based on their extensions.
6. **Source Scan Error Policy:** By default, an unreadable source directory stops the operation. With `-ignore-directory-errors`, LookThrough skips inaccessible source directories, records each unique skipped path, and reports the paths and errors when the operation ends.
7. **Analysis Reporting:** After processing, LookThrough reports the number of source files scanned, the number of unique files copied during the current run, and the total logical size of duplicate data omitted from the output.

Deduplication is based on SHA-256 hashes. LookThrough does not perform an additional byte-for-byte comparison after two files produce the same hash.

---

<div id="features"></div>

## ✨ Features

- **Bounded Worker Pool:** Uses a fixed number of workers instead of creating one goroutine per file, reducing scheduler and I/O contention.
- **Automatic Worker Selection:** Uses `GOMAXPROCS × 2`, limited to 4–32 workers, unless overridden with `-j`.
- **Concurrent Metadata Lookup:** Keeps `WalkDir` sequential while moving file metadata lookup into the worker pool.
- **Fixed-Total Progress Tracking:** Counts eligible files before processing, then uses an `atomic.Int64` processed counter while a single timed goroutine renders progress against the fixed total.
- **Timed Progress Indicator:** Renders the latest processed count at most once every 100 milliseconds, avoiding per-file terminal writes while still showing intermediate values.
- **Optional In-Memory Consistency Check:** With `-verify`, cross-checks the hashes recorded during processing in linear time. It does not reopen or rehash the copied destination files.
- **Optional Safe Copy Mode:** With `-safe-copy`, writes each unique file to an internal temporary path and renames it to the final destination only after a successful copy and close. Direct copying remains the faster default.
- **Configurable Source Scan Errors:** Unreadable source directories fail the operation by default. With `-ignore-directory-errors`, inaccessible directories are skipped and included in a deduplicated final warning report.
- **Bounded In-Flight Work:** Worker count and channel capacity limit simultaneous file processing. Deduplication maps and optional verification lists still grow proportionally to the number of files discovered.
- **Flexible Input Methods:** Accepts a path through `-p`, as a positional argument, or through an interactive prompt.
- **Human-Readable Output:** Converts byte counts into B, KB, MB, GB, or TB.
- **File Categorization:** With `-s`, organizes files into folders such as `Documents`, `Images`, `Electronics_Automation`, and `Programming`.
- **Existing Output Handling:** If the output directory already exists, prompts the user to delete its contents or amend it while preserving existing files.

---

<div id="requirements"></div>

## 📋 Requirements

- **Go Compiler:** Go 1.19 or newer.
- **Operating System:** Designed for Windows, macOS, and Linux.

---

<div id="installation"></div>

## 📦 Installation

Clone the repository and build the executable with the Go toolchain:

```sh
git clone https://github.com/LeonardoCG12/LookThrough.git
cd LookThrough
go build -o lookthrough ./src
```

On Windows, you may build an `.exe` file instead:

```powershell
go build -o lookthrough.exe ./src
```

<div id="usage"></div>

## 🚀 Usage

LookThrough accepts the target directory through command-line flags, a positional argument, or an interactive prompt.

<div id="using-cli-flags"></div>

### 🧰 1. Using CLI Flags (Recommended)

Provide the target directory and enable optional features with flags:

```sh
./lookthrough -p /absolute/path/to/target/directory -b -s -j 8 -verify -safe-copy -ignore-directory-errors
```

- `-p`: Target directory to scan.
- `-b`: Counts eligible files first, then displays a live processed-versus-total progress indicator.
- `-s`: Sorts copied files into category directories based on extension.
- `-j`: Number of concurrent workers. Use `0` or omit the flag for automatic selection. Manual values from 1 to 128 are accepted.
- `-verify`: Enables the optional in-memory hash-list consistency check. It does not verify copied files by reading them again from disk.
- `-safe-copy`: Writes each destination through a temporary file and renames it only after copying and closing succeed. This better prevents incomplete files from appearing under their final names, but is slower for workloads containing many small files.
- `-ignore-directory-errors`: Skips unreadable source directories instead of stopping immediately. Every unique skipped directory and its access error are reported when the operation ends. Errors involving source files, the output directory, hashing, or copying still fail the operation.

The automatic worker count is calculated as twice GOMAXPROCS, with a minimum of 4 and a maximum of 32.

A manual value overrides the automatic calculation. Values below 0 or above 128 are rejected instead of being silently reduced.

> 🤖 **Automation note:** If the output directory already exists, LookThrough prompts for a delete or amend decision even when CLI flags are used. Unattended scripts should ensure that the output directory does not already exist.

<div id="passing-the-directory-directly"></div>

### 📂 2. Passing the Directory Directly

Provide the target directory as the first positional argument:

```sh
./lookthrough /absolute/path/to/target/directory
```

When no optional flags are provided, LookThrough uses automatic worker selection, direct copying, strict source-directory error handling, and leaves progress display, categorization, and verification disabled.

<div id="interactive-mode"></div>

### ⌨️ 3. Interactive Mode

Run the executable without `-p` or a positional directory to enter the path interactively:

```sh
./lookthrough
Choose a directory to look through: /absolute/path/to/target/directory
```

For a target named `MyFolder`, the output directory is created as `MyFolder/new-MyFolder`. If that directory already exists, LookThrough asks how to proceed:

```text
[!] The folder '/absolute/path/to/MyFolder/new-MyFolder' already exists.
[?] Do you want to delete its content or amend it? (d/a):
```

- `d` or `delete`: Removes the existing output directory and creates a new one.
- `a` or `amend`: Keeps the existing output files, loads their hashes, and adds only content that is not already present. Internal temporary files left by an interrupted `-safe-copy` run are ignored. The final summary reports files copied during the current run rather than the total number of files already present in the output directory.

<div id="file-categorization-summary"></div>

## 🗂️ File Categorization Summary

When using `-s`, LookThrough organizes files into predefined categories. Examples include:

- **Images:** `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.svg`, and other image formats.
- **Videos:** `.mp4`, `.mov`, `.avi`, `.mkv`, `.wmv`, and other video formats.
- **Audio:** `.mp3`, `.wav`, `.flac`, `.aac`, `.ogg`, and other audio formats.
- **Electronics_Automation:** `.ino`, `.hex`, `.plc`, `.l5x`, `.awl`, and related engineering formats.
- **Design_3D_CAD:** `.step`, `.stl`, `.obj`, `.sldprt`, and other CAD or 3D formats.
- **Programming:** `.go`, `.py`, `.c`, `.cpp`, `.rs`, `.js`, and other source or configuration formats.
- **Others:** Files whose extensions do not match a predefined category.

<div id="system-exclusions"></div>

## 🚫 System Exclusions

LookThrough skips the following filenames case-insensitively:

- `desktop.ini`
- `thumbs.db`
- `.DS_Store`

To prevent recursive processing, it also excludes the exact output directory created for the current target and all of that directory's descendants. It does not automatically exclude every unrelated directory whose name begins with `new-`.

### ⚠️ Source Directory Access Errors

By default, LookThrough stops when it cannot access a source directory. This prevents a successful result from silently representing only part of the requested source tree.

With `-ignore-directory-errors`, inaccessible source directories are skipped. Because the source tree is walked once for counting and again for processing, skipped paths are deduplicated before being displayed. A successful operation with skipped directories prints a warning report such as:

```text
[!] SKIPPED SOURCE DIRECTORIES: 2
[!] /absolute/path/private: permission denied
[!] /absolute/path/restricted: access is denied
```

The result is necessarily incomplete with respect to those directories. The flag applies only to source-directory traversal errors; errors involving individual files or the destination remain fatal.

---

<div id="sample-output"></div>

## 🖥️ Sample Output

A successful execution currently prints output similar to:

```text
[+] SUCCESS
[+] UNIQUE FILES HAVE BEEN COPIED

>>> Source Files Scanned: 1245
>>> Unique Files Copied This Run: 912
>>> Duplicate Data Skipped: 4.2GB
```

`Duplicate Data Skipped` is the total logical size reported by the source files that were not copied because their SHA-256 hashes were already present. LookThrough does not delete the source files, so this value is not disk space physically reclaimed from the source directory.

If no eligible files are found:

```text
[-] FAIL
[-] NO FILES FOUND

[!] CRITICAL ERROR: no files found
```

If the optional in-memory verification fails:

```text
[+] VERIFYING FILE HASHES...

[-] FAIL
[-] IN-MEMORY HASH CONSISTENCY CHECK FAILED

[!] CRITICAL ERROR: in-memory hash consistency check failed
```

This verification failure reports an inconsistency between hashes recorded in memory. It does not mean that LookThrough reopened and rehashed every destination file.

Other operational failures are reported as:

```text
[!] CRITICAL ERROR: <error details>
```

<p align="right">[<a href="#top">Back to top</a>]</p>
