# Score Checker

[![CI](https://github.com/mcreekmore/score-checker/actions/workflows/ci.yml/badge.svg)](https://github.com/mcreekmore/score-checker/actions/workflows/ci.yml)
[![Docker](https://github.com/mcreekmore/score-checker/actions/workflows/docker.yml/badge.svg)](https://github.com/mcreekmore/score-checker/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mcreekmroe/score-checker)](https://goreportcard.com/report/github.com/mcreekmore/score-checker)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A microservice that monitors Sonarr episodes and Radarr movies for low custom format scores and optionally triggers automatic searches for better quality versions.

## Features

- **Multi-Service Support**: Works with both Sonarr (TV shows) and Radarr (movies)
- **Multiple Instances**: Support for multiple Sonarr and Radarr instances per application
- **Batch Processing**: Process a configurable number of items per run to avoid overwhelming your system
- **Scheduled Execution**: Run as a daemon with configurable intervals (e.g., every hour)
- **Flexible Configuration**: Support for config files, environment variables, and command-line flags
- **Docker Ready**: Containerized deployment with proper configuration management
- **Safe Operation**: Dry-run mode by default - only reports findings unless explicitly enabled

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/mcreekmore/score-checker/releases).

```bash
# Linux
wget https://github.com/mcreekmore/score-checker/releases/latest/download/score-checker-linux-amd64
chmod +x score-checker-linux-amd64
./score-checker-linux-amd64 --help

# macOS
wget https://github.com/mcreekmore/score-checker/releases/latest/download/score-checker-darwin-amd64
chmod +x score-checker-darwin-amd64
./score-checker-darwin-amd64 --help
```

### Docker

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/mcreekmore/score-checker:latest

# Or build locally
docker build -t score-checker .
```

### From Source

```bash
git clone https://github.com/mcreekmore/score-checker.git
cd score-checker
make build
```

## Configuration

Configuration can be provided via:
1. Command-line flags
2. Environment variables (prefixed with `SCORECHECK_`)
3. Configuration file (`config.yaml`)

### Configuration Options

| Option         | Flag              | Environment                | Default | Description                                 |
| -------------- | ----------------- | -------------------------- | ------- | ------------------------------------------- |
| Trigger Search | `--triggersearch` | `SCORECHECK_TRIGGERSEARCH` | `false` | Actually trigger searches (vs. report only) |
| Batch Size     | `--batchsize`     | `SCORECHECK_BATCHSIZE`     | `5`     | Items to check per run                      |
| Interval       | `--interval`      | `SCORECHECK_INTERVAL`      | `1h`    | Daemon mode interval                        |
| Log Level      | `--loglevel`      | `SCORECHECK_LOGLEVEL`      | `INFO`  | Logging verbosity (ERROR, INFO, DEBUG, VERBOSE) |

**Note**: Sonarr and Radarr instances are configured via the config file only (see below).

### Configuration File

Create a `config.yaml` file:

```yaml
# Sonarr instances - array of instances, each with a name, baseurl, and apikey
sonarr:
  - name: "main"
    baseurl: "http://localhost:8989"
    apikey: "your-sonarr-api-key-here"
  - name: "4k"
    baseurl: "http://localhost:8990"
    apikey: "your-4k-sonarr-api-key-here"

# Radarr instances - array of instances, each with a name, baseurl, and apikey
radarr:
  - name: "main"
    baseurl: "http://localhost:7878"
    apikey: "your-radarr-api-key-here"
  - name: "4k"
    baseurl: "http://localhost:7879"
    apikey: "your-4k-radarr-api-key-here"

# General settings
triggersearch: false
batchsize: 5
interval: "1h"

# Logging level - controls output verbosity
# ERROR:   Only errors
# INFO:    Errors and info messages (default)
# DEBUG:   Errors, info, and debug messages
# VERBOSE: All messages including detailed output
loglevel: "INFO"
```

**Multiple Instances**: You can configure multiple Sonarr and/or Radarr instances by adding more entries to the respective arrays. Each instance must have a unique name, baseurl, and apikey.

## Usage

### Run Once

Check items once and exit:

```bash
# Check all configured instances
./score-checker

# Process 10 items and trigger searches
./score-checker --batchsize 10 --triggersearch
```

### Daemon Mode

Run continuously with scheduled checks:

```bash
# Run daemon with config file
./score-checker daemon

# Run daemon with custom interval
./score-checker daemon --interval 30m
```

### Docker

```bash
# Run once with config file
docker run --rm -v $(pwd)/config.yaml:/app/config.yaml ghcr.io/mcreekmore/score-checker:latest

# Run as daemon
docker run -d --restart unless-stopped \
  -v $(pwd)/config.yaml:/app/config.yaml \
  --name score-checker \
  ghcr.io/mcreekmore/score-checker:latest daemon

# Run with environment variables
docker run -d --restart unless-stopped \
  -e SCORECHECK_BATCHSIZE=10 \
  -e SCORECHECK_INTERVAL=2h \
  -e SCORECHECK_TRIGGERSEARCH=true \
  -v $(pwd)/config.yaml:/app/config.yaml \
  --name score-checker \
  ghcr.io/mcreekmore/score-checker:latest daemon
```

### Docker Compose

```bash
# Use the included docker-compose.yml
docker-compose up -d

# For one-time run
docker-compose --profile once up score-checker-once

# For environment-based config
docker-compose --profile env-config up -d score-checker-env
```

## Examples

### Basic Usage

```bash
# Check 5 items from all configured instances, report only
./score-checker

# Check 10 items and trigger searches
./score-checker --batchsize 10 --triggersearch
```

### Daemon Examples

```bash
# Check every 30 minutes
./score-checker daemon --interval 30m

# Check every 2 hours, process 20 items per run
./score-checker daemon --interval 2h --batchsize 20
```

### Environment Variables

```bash
export SCORECHECK_BATCHSIZE=15
export SCORECHECK_INTERVAL=90m
export SCORECHECK_TRIGGERSEARCH=true

# Run daemon with environment config (requires config file for instances)
./score-checker daemon
```

**Note**: Multiple instances must be configured via config file. Environment variables only support the general settings.

## How It Works

### Sonarr (TV Shows)
1. **Fetches Series**: Retrieves all series from your Sonarr instance
2. **Checks Episodes**: Examines episodes with files for custom format scores
3. **Identifies Low Scores**: Finds episodes with scores below zero
4. **Batch Processing**: Processes only the configured number of episodes per run
5. **Optional Search**: Triggers automatic searches if enabled

### Radarr (Movies)
1. **Fetches Movies**: Retrieves all movies from your Radarr instance
2. **Checks Movies**: Examines movies with files for custom format scores
3. **Identifies Low Scores**: Finds movies with scores below zero
4. **Batch Processing**: Processes only the configured number of movies per run
5. **Optional Search**: Triggers automatic searches if enabled

Both services provide detailed reporting of findings.

## Output Example

```
Search triggering is ENABLED - will automatically search for better versions
Batch size: 5 items per run
Found 2 Sonarr instance(s)

=== Checking Sonarr Instance: main ===
[main] Fetching series and checking custom format scores...
[main] Checking series: Breaking Bad (ID: 1)
[main] Checking series: The Office (ID: 2)
[main] Reached batch limit of 5 episodes

[main] Found 2 episode(s) with custom format scores below zero:
[main] (Searches have been triggered for these episodes)

[main] Series: Breaking Bad
[main]   Episode: S01E01 - Pilot
[main]   Custom Format Score: -10
[main]   Episode ID: 123

=== Checking Sonarr Instance: 4k ===
[4k] Fetching series and checking custom format scores...
[4k] No episodes found with custom format scores below zero.

Found 1 Radarr instance(s)

=== Checking Radarr Instance: main ===
[main] Fetching movies and checking custom format scores...
[main] Checking movie: The Matrix (1999)
[main] Checking movie: Inception (2010)

[main] Found 1 movie(s) with custom format scores below zero:
[main] (Searches have been triggered for these movies)

[main] Movie: The Matrix (1999)
[main]   Custom Format Score: -15
[main]   Movie ID: 789
```

## Safety Features

- **Dry-run by default**: Only reports findings unless `--triggersearch` is enabled
- **Batch limiting**: Prevents overwhelming your system with too many simultaneous operations
- **Error handling**: Continues processing other series if one fails
- **Configurable intervals**: Prevents excessive API calls

## Development

### Building and Testing

```bash
# Build the application
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Generate HTML coverage report
make test-coverage-html

# Run all quality checks
make quality

# Build Docker image
make docker-build

# Build release binaries for all platforms
make build-release
```

### CI/CD

This project uses GitHub Actions for continuous integration and deployment:

- **CI Pipeline** (`.github/workflows/ci.yml`):
  - Runs on pushes to `main` and `develop` branches and pull requests
  - Tests across Go versions with race detection
  - Runs linting with golangci-lint
  - Performs security scanning with Gosec
  - Enforces code coverage thresholds (70%+)
  - Builds binaries for multiple platforms

- **Docker Pipeline** (`.github/workflows/docker.yml`):
  - Builds multi-platform Docker images (linux/amd64, linux/arm64)
  - Publishes to GitHub Container Registry (GHCR)
  - Generates Software Bill of Materials (SBOM)
  - Performs vulnerability scanning
  - Tags images based on Git tags and branches

- **Release Pipeline** (`.github/workflows/release.yml`):
  - Triggers on Git tags (e.g., `v1.0.0`)
  - Creates GitHub releases with changelogs
  - Builds and uploads platform-specific binaries
  - Generates checksums for verification

### Security

- **Automated dependency updates** via Dependabot
- **Vulnerability scanning** in CI/CD pipeline
- **Security scanning** with Gosec static analyzer
- **Minimal Docker images** using scratch base for reduced attack surface
- **Non-root container execution** for improved security

## Requirements

- Go 1.24.4+ (for building from source)
- Docker (for containerized deployment)
- Access to Sonarr and/or Radarr instances with API enabled
- Valid API keys for the services you want to monitor

## License

MIT License