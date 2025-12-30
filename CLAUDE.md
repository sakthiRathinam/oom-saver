# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OOM-saver is a Go-based system process monitoring tool designed to prevent out-of-memory situations. Currently in early development with Linux-focused implementation.

## Development Commands

### Build
```bash
go build -o oom-saver main.go
```

### Run
```bash
go run main.go
```

### Testing
```bash
go test ./...
```

### Format Code
```bash
go fmt ./...
```

### Run Linter (if golangci-lint is installed)
```bash
golangci-lint run
```

## Code Architecture

### Current Structure

The codebase is currently minimal with a single-file architecture in `main.go`:

- **Process struct** (`main.go:19-23`): Core data structure representing system processes with Name, PID, and Status fields
- **get_all_running_processes_from_os()** (`main.go:25-37`): OS detection layer that routes to platform-specific implementations based on runtime.GOOS

### Platform Support

The architecture uses runtime.GOOS to detect the operating system:
- **Linux**: Partially implemented (calls `get_all_running_processes_from_linux()` but function is not yet defined)
- **Windows**: Not implemented (returns error)
- **macOS (darwin)**: Not implemented (returns error)

### Known Issues

- The `get_all_running_processes_from_linux()` function is called in `main.go:31` but not yet implemented, which will cause compilation errors
- The function at `main.go:25-37` always returns an error on line 36, even for Linux, due to the missing implementation

### Development Notes

When adding platform-specific implementations, follow the pattern of creating separate functions for each OS (e.g., `get_all_running_processes_from_linux()`, `get_all_running_processes_from_windows()`, etc.) and routing through the main `get_all_running_processes_from_os()` function.
