# CI/CD Setup Guide

This document explains the complete CI/CD setup for the score-checker project using GitHub Actions.

## Overview

The project uses a comprehensive CI/CD pipeline with three main workflows:

1. **Continuous Integration** (`ci.yml`) - Testing, linting, and quality checks
2. **Docker Build and Publish** (`docker.yml`) - Container image creation and publishing
3. **Release Management** (`release.yml`) - Automated releases with multi-platform binaries

## Workflows

### 1. Continuous Integration (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main` and `develop` branches
- Pull requests to `main` branch

**Jobs:**

#### Test Job
- Runs on `ubuntu-latest`
- Sets up Go 1.24
- Caches Go modules for faster builds
- Downloads and verifies dependencies
- Runs `go vet` for static analysis
- Checks code formatting with `gofmt`
- Executes tests with race detection and coverage
- Generates HTML coverage report
- Enforces 70% minimum coverage threshold
- Uploads coverage artifacts

#### Lint Job
- Uses `golangci-lint` for comprehensive linting
- Runs in parallel with test job
- 5-minute timeout for efficiency

#### Build Job
- Depends on successful test and lint jobs
- Builds binaries for multiple platforms:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64)
- Uploads build artifacts

#### Security Job
- Runs Nancy vulnerability scanner for dependencies
- Runs govulncheck for Go vulnerability database scanning
- Identifies potential security vulnerabilities in dependencies and stdlib
- Runs independently for faster feedback

### 2. Docker Build and Publish (`.github/workflows/docker.yml`)

**Triggers:**
- Push to `main` branch
- Git tags matching `v*` pattern
- Pull requests to `main` (build only, no push)

**Features:**
- Multi-platform builds (linux/amd64, linux/arm64)
- Publishes to GitHub Container Registry (GHCR)
- Automatic tagging strategy:
  - `latest` for main branch
  - Version tags for releases (e.g., `v1.0.0`, `v1.0`, `v1`)
  - Branch names for feature branches
- Build cache optimization
- Software Bill of Materials (SBOM) generation
- Vulnerability scanning with Anchore
- SARIF results upload for security dashboard

### 3. Release Management (`.github/workflows/release.yml`)

**Triggers:**
- Git tags matching `v*` pattern (e.g., `v1.0.0`)

**Process:**
1. Runs full test suite
2. Builds release binaries for all platforms
3. Generates SHA256 checksums
4. Creates automatic changelog from git commits
5. Creates GitHub release with:
   - Release notes
   - Platform-specific binaries
   - Checksum file
   - Docker image links
6. Marks pre-releases for tags containing `-` (e.g., `v1.0.0-beta`)

## Security Features

### Automated Dependency Updates
- **Dependabot** configuration (`.github/dependabot.yml`)
- Weekly updates for:
  - Go modules
  - GitHub Actions
  - Docker base images
- Automatic PR creation with security patch information

### Security Scanning
- **Nancy** dependency vulnerability scanning
- **govulncheck** Go vulnerability database scanning
- **Anchore** container vulnerability scanning
- **SARIF** results integration with GitHub Security tab
- **SBOM** generation for supply chain transparency

### Container Security
- Multi-stage Docker builds
- Minimal `scratch` base image
- Non-root user execution
- Static binary with no dependencies
- CA certificates and timezone data included

## Configuration Requirements

### Repository Settings

#### Secrets (if needed)
- No additional secrets required - uses built-in `GITHUB_TOKEN`

#### Permissions
- Repository needs `packages: write` permission for GHCR
- Actions need `contents: write` for releases
- Security events need `security-events: write` for SARIF upload

#### Branch Protection (Recommended)
```yaml
# .github/branch-protection.yml
rules:
  main:
    required_status_checks:
      strict: true
      contexts:
        - "Test"
        - "Lint" 
        - "Build"
        - "Security Scan"
    enforce_admins: true
    required_pull_request_reviews:
      required_approving_review_count: 1
      dismiss_stale_reviews: true
    restrictions: null
```

### Environment Variables

#### For Development
```bash
GO_VERSION="1.24"
REGISTRY="ghcr.io"
IMAGE_NAME="${{ github.repository }}"
```

#### For Runtime (Docker)
```bash
SCORECHECK_TRIGGERSEARCH=false
SCORECHECK_BATCHSIZE=5
SCORECHECK_INTERVAL=1h
```

## Usage Examples

### Creating a Release

1. **Create and push a tag:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **The release workflow will:**
   - Build binaries for all platforms
   - Create GitHub release
   - Generate checksums
   - Build and push Docker images with version tags

### Docker Image Usage

```bash
# Pull latest image
docker pull ghcr.io/yourusername/score-checker:latest

# Pull specific version
docker pull ghcr.io/yourusername/score-checker:v1.0.0

# Run with config file
docker run --rm -v $(pwd)/config.yaml:/app/config.yaml \
  ghcr.io/yourusername/score-checker:latest

# Run as daemon
docker run -d --restart unless-stopped \
  -v $(pwd)/config.yaml:/app/config.yaml \
  --name score-checker \
  ghcr.io/yourusername/score-checker:latest daemon
```

### Development Workflow

```bash
# Run quality checks locally
make quality

# Build Docker image
make docker-build

# Test Docker image
make docker-run

# Build release binaries
make build-release
```

## Monitoring and Maintenance

### Build Status
- Monitor workflow status on GitHub Actions tab
- Set up notifications for failed builds
- Review security scan results regularly

### Dependency Updates
- Review and merge Dependabot PRs
- Test dependency updates in staging environment
- Monitor for breaking changes

### Performance Monitoring
- Track build times and optimize as needed
- Monitor Docker image sizes
- Review test execution times

## Troubleshooting

### Common Issues

#### Go Version Mismatch
- Ensure Dockerfile and workflows use Go 1.24+
- Update `go.mod` if needed

#### Docker Build Failures
- Check `.dockerignore` is properly configured
- Verify multi-stage build steps
- Test locally with `make docker-build`

#### Test Failures
- Run tests locally with `make test`
- Check coverage thresholds
- Review race condition detection

#### Security Scan Failures
- Review Nancy and govulncheck findings
- Update vulnerable dependencies
- Check container scan results

### Getting Help
- Check GitHub Actions logs for detailed error messages
- Review workflow files for configuration issues
- Test changes in feature branches before merging

## Future Enhancements

### Potential Improvements
1. **Integration Testing**: Add end-to-end tests with real Sonarr/Radarr instances
2. **Performance Testing**: Add benchmark comparisons in CI
3. **Multi-Environment Deployment**: Add staging/production deployment workflows
4. **Helm Charts**: Add Kubernetes deployment manifests
5. **Code Quality Gates**: Integrate additional quality metrics
6. **Notification System**: Add Slack/Discord notifications for releases

### Monitoring Additions
1. **Metrics Collection**: Add Prometheus metrics endpoint
2. **Health Checks**: Enhanced health check endpoints
3. **Distributed Tracing**: Add OpenTelemetry support
4. **Log Aggregation**: Structured logging for better observability