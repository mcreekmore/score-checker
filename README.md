# Score Checker

[![CI](https://github.com/mcreekmore/score-checker/actions/workflows/ci.yml/badge.svg)](https://github.com/mcreekmore/score-checker/actions/workflows/ci.yml)
[![Docker](https://github.com/mcreekmore/score-checker/actions/workflows/docker.yml/badge.svg)](https://github.com/mcreekmore/score-checker/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mcreekmore/score-checker)](https://goreportcard.com/report/github.com/mcreekmore/score-checker)
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