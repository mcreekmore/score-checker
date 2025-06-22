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

| Option         | Flag              | Environment                | Default | Description                                     |
| -------------- | ----------------- | -------------------------- | ------- | ----------------------------------------------- |
| Trigger Search | `--triggersearch` | `SCORECHECK_TRIGGERSEARCH` | `false` | Actually trigger searches (vs. report only)     |
| Batch Size     | `--batchsize`     | `SCORECHECK_BATCHSIZE`     | `5`     | Items to check per run                          |
| Interval       | `--interval`      | `SCORECHECK_INTERVAL`      | `1h`    | Daemon mode interval                            |
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
loglevel: "INFO" # INFO, ERROR, DEBUG, VERBOSE
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

```yaml
---
services:
  score-checker:
    image: ghcr.io/mcreekmore/score-checker:latest
    container_name: score-checker
    restart: unless-stopped
    volumes:
      - /path/to/score-checker/data:/etc/score-checker
    environment:
      - SCORECHECK_TRIGGERSEARCH=false
      - SCORECHECK_BATCHSIZE=5
      - SCORECHECK_INTERVAL=1h
    command: ["daemon"]
```

**Note**: Multiple instances must be configured via config file. Environment variables only support the general settings.

## Requirements

- Go 1.24.4+ (for building from source)
- Docker (for containerized deployment)
- Access to Sonarr and/or Radarr instances with API enabled
- Valid API keys for the services you want to monitor

## License

MIT License